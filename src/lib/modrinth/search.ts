import type { Pack } from "../pack";
import { selectFromList, statusMessage } from "../util";
import { BaseURL } from "./apiUrl";
import chalk from "chalk";

export class ModrinthProject {
    slug: string;
    title: string;
    description: string;
    categoies: string[];
    client_side: "required" | "optional" | "unsupported" | "unknown";
    server_side: "required" | "optional" | "unsupported" | "unknown";
    body: string;
    status: "approved" | "archived" | "rejected" | "draft" | "unlisted" | "processing" | "wildheld" | "scheduled" | "private" | "unknown";
    requested_status: "approved" | "archived" | "unlisted" | "private" | "draft";
    additional_categories: string[];
    issues_url: string;
    source_url: string;
    wiki_url: string;
    discord_url: string;
    donation_url: string;
    project_type: "mod" | "modpack" | "resourcepack" | "shader";
    downloads: number;
    color?: number;
    thread_id?: string;
    monetization_status: "monetized" | "demonetized" | "force-demonetized";
    id: string;
    team: string;
    body_url?: string; // "always null"...?
    moderator_message?: {
        message?: string;
        body?: string;
    }
    published: Date;
    updated: Date;
    approved?: Date;
    queued?: Date;
    followers: number;
    liscence: {
        id: string;
        name: string;
        url?: string;
    }
    versions: string[];
    game_versions: string[];
    loaders: string[];
    gallery: {
        url: string;
        featured: boolean;
        title?: string;
        description?: string;
        created: Date;
        ordering: number;
    }[];

    constructor(data: any) {
        this.slug = data.slug;
        this.title = data.title;
        this.description = data.description;
        this.categoies = data.categories;
        this.client_side = data.client_side;
        this.server_side = data.server_side;
        this.body = data.body;
        this.status = data.status;
        this.requested_status = data.requested_status;
        this.additional_categories = data.additional_categories || [];
        this.issues_url = data.issues_url || "";
        this.source_url = data.source_url || "";
        this.wiki_url = data.wiki_url || "";
        this.discord_url = data.discord_url || "";
        this.donation_url = data.donation_url || "";
        this.project_type = data.project_type;
        this.downloads = data.downloads;
        this.color = data.color;
        this.thread_id = data.thread_id;
        this.monetization_status = data.monetization_status;
        this.id = data.id;
        this.team = data.team;
        this.body_url = data.body_url || null; // "always null" according to the API docs
        this.moderator_message = {
            message: data.moderator_message?.message,
            body: data.moderator_message?.body || null
        };
        this.published = new Date(data.published);
        this.updated = new Date(data.updated);
        if (data.approved) {
            this.approved = new Date(data.approved);
        }
        if (data.queued) {
            this.queued = new Date(data.queued);
        }
        this.followers = data.followers;
        this.liscence = {
            id: data.license.id,
            name: data.license.name,
            url: data.license.url || ""
        };
        this.versions = data.versions;
        this.game_versions = data.game_versions;
        this.loaders = data.loaders;
        if (data.gallery) {
            this.gallery = data.gallery.map((item: any) => ({
                url: item.url,
                featured: item.featured,
                title: item.title,
                description: item.description,
                created: new Date(item.created),
                ordering: item.ordering
            }));
        } else {
            this.gallery = [];
        }
    }
}

export async function findProjectByIdOrSlug(id: string, packData?: Pack, verbose: boolean = false): Promise<ModrinthProject | null> {
    if (verbose) {
        console.log(chalk.gray(`Starting search for project with ID or slug: ${chalk.yellowBright(id)}`));
        if (packData) {
            console.log(chalk.gray(`Pack data provided: gameVersion=${packData.gameVersion}, modloader=${packData.modloader.name}`));
        }
    }
    const status = statusMessage(chalk.gray(`Looking for project with ID or slug: ${chalk.yellowBright(id)}`), { clearLine: true });
    const res = await fetch(`${BaseURL}/project/${id}`)
    status.clear();
    if (res.ok) {
        if (verbose) {
            console.log(chalk.gray(`Received successful response from API for ${id}`));
        }
        const data = await res.json();
        const project = new ModrinthProject(data);
        if (!packData || (project.game_versions.includes(packData.gameVersion) && project.loaders.includes(packData.modloader.name))) {
            if (verbose) {
                console.log(chalk.gray(`Found project: ${project.title} (${project.slug})`));
            }
            return project;
        } else {
            if (verbose) {
                console.log(chalk.gray(`Project ${project.title} (${project.slug}) does not match the pack's game version or modloader.`));
                console.log(chalk.gray(`Project game_versions: ${project.game_versions.join(", ")}, loaders: ${project.loaders.join(", ")}`));
            }
            return null;
        }
    } else {
        if (verbose) {
            console.log(chalk.red(`Failed to fetch project: ${id}. Status: ${res.status}`));
        }
        return null;
    }
}

export async function findProjectBySearch(query: string, packData?: Pack, verbose: boolean = false): Promise<ModrinthProject | null> {
    if (verbose) {
        console.log(chalk.gray(`Searching for project with query: ${chalk.yellowBright(query)}`));
        if (packData) {
            console.log(chalk.gray(`Pack data provided: gameVersion=${packData.gameVersion}, modloader=${packData.modloader.name}`));
        }
    }
    let facets = "";
    if (packData) {
        facets = `&facets=[[\"versions:${packData.gameVersion}\"],[\"project_type:mod\",\"project_type:resourcepack\",\"project_type:shader\"],[\"categories:${packData.modloader.name}\"]]`;
        if (verbose) {
            console.log(chalk.gray(`Using facets: ${facets}`));
        }
    }
    const status = statusMessage(chalk.gray(`Searching for project with query: ${chalk.yellowBright(query)}`), { clearLine: true });
    const res = await fetch(`${BaseURL}/search?query=${encodeURIComponent(query)}${facets}&limit=5`);
    status.clear();
    if (res.ok) {
        if (verbose) {
            console.log(chalk.gray(`Received successful response from API for query: ${query}`));
        }
        const data = await res.json();
        if (data.hits && data.hits.length > 1) {
            if (verbose) {
                console.log(chalk.gray(`Multiple results found (${data.hits.length}). Formatting and prompting user selection.`));
            }
            const formattedData: ModrinthProject[] = await Promise.all(data.hits.map((item: any) => findProjectByIdOrSlug(item.slug, packData, verbose)));
            const list: Record<string, string> = {};
            formattedData.forEach(project => {
                list[project.id] = `${chalk.green.bold(project.title)} ${chalk.gray("(" + project.slug + ")")}\n${"    " + project.description}`;
            });
            const selectList = await selectFromList(
                Object.values(list),
                chalk.blueBright.bold(`Multiple results found. Select from the top ${Object.values(list).length} results: `)
            );
            // Only filter for return value if packData is present
            const matchingProjects = packData
                ? formattedData.filter(project => 
                    project.game_versions.includes(packData.gameVersion) && 
                    project.loaders.includes(packData.modloader.name)
                )
                : formattedData;
            // Return the selected project if found, else first matching or null
            const selectedProject = formattedData[selectList] || matchingProjects[0] || null;
            if (verbose && selectedProject) {
                console.log(chalk.gray(`Selected project: ${selectedProject.title} (${selectedProject.slug})`));
            }
            return selectedProject;
        } else if (data.hits && data.hits.length === 1) {
            if (verbose) {
                console.log(chalk.gray(`Single result found. Returning project: ${data.hits[0].title} (${data.hits[0].slug})`));
            }
            const project = new ModrinthProject(data.hits[0]);
            return project;
        } else {
            if (verbose) {
                console.log(chalk.gray(`No results found for query: ${query}`));
            }
            return null;
        }
    } else {
        if (verbose) {
            console.log(chalk.red(`Failed to fetch search results for query: ${query}. Status: ${res.status}`));
        }
        return null;
    }
}

export async function findProject(query: string, packData?: Pack, verbose: boolean = false): Promise<ModrinthProject | null> {
    if (verbose) {
        console.log(chalk.gray(`Attempting to find project by ID or slug: ${query}`));
    }
    const idSearchResult = await findProjectByIdOrSlug(query, packData, verbose);
    if (idSearchResult) {
        if (verbose) {
            console.log(chalk.gray(`Project found by ID or slug: ${idSearchResult.title} (${idSearchResult.slug})`));
        }
        return idSearchResult;
    } else {
        if (verbose) {
            console.log(chalk.gray(`No project found by ID or slug. Searching by query: ${query}`));
        }
        const searchResult = await findProjectBySearch(query, packData, verbose);
        if (searchResult) {
            if (verbose) {
                console.log(chalk.gray(`Project found by search: ${searchResult.title} (${searchResult.slug})`));
            }
            return searchResult;
        } else {
            if (verbose) {
                console.log(chalk.gray(`No project found by search for query: ${query}`));
            }
            return null;
        }
    }
}