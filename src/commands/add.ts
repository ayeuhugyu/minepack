import chalk from "chalk";
import path from "path";
import fs from "fs-extra";
import { Command, registerCommand } from "../lib/command";
import { Mod, ModSide, HashFormat, type ModData } from "../lib/mod";

const MODRINTH_API = "https://api.modrinth.com/v2";

async function searchModrinth(query: string) {
    const url = `${MODRINTH_API}/search?query=${encodeURIComponent(query)}&limit=5`;
    const res = await fetch(url);
    if (!res.ok) throw new Error("Failed to search Modrinth");
    const data = await res.json();
    return data.hits || [];
}

async function getModrinthProject(idOrSlug: string) {
    const url = `${MODRINTH_API}/project/${idOrSlug}`;
    const res = await fetch(url);
    if (!res.ok) return null;
    return await res.json();
}

async function getModrinthVersion(projectId: string, gameVersion?: string, loader?: string) {
    let url = `${MODRINTH_API}/project/${projectId}/version`;
    const params = [];
    if (gameVersion) params.push(`game_versions=[\"${gameVersion}\"]`);
    if (loader) params.push(`loaders=[\"${loader}\"]`);
    if (params.length) url += `?${params.join("&")}`;
    const res = await fetch(url);
    if (!res.ok) return null;
    const versions = await res.json();
    return versions[0] || null;
}

async function getModrinthFileCdnUrl(version: any) {
    if (!version || !version.files || !version.files.length) return null;
    // Prefer primary file
    const file = version.files.find((f: any) => f.primary) || version.files[0];
    return file.url;
}

const addCommand = new Command({
    name: "add",
    description: "Add a mod to your modpack from Modrinth or a direct URL.",
    arguments: [
        {
            name: "mod",
            aliases: [],
            description: "Modrinth URL, mod ID, or search term.",
            required: true
        }
    ],
    flags: [
        {
            name: "download",
            aliases: ["D"],
            description: "Download the mod jar directly instead of creating a .json stub.",
            takesValue: false
        },
        {
            name: "side",
            aliases: ["s"],
            description: "Which side to use this mod on (client/server/both)",
            takesValue: true
        },
        {
            name: "url",
            aliases: [],
            description: "Direct download URL (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "name",
            aliases: [],
            description: "Name of the mod (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "filename",
            aliases: [],
            description: "Filename for the mod (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "hash",
            aliases: [],
            description: "Hash for the mod file (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "hash-format",
            aliases: [],
            description: "Hash format (sha1, sha256, etc) (if not using Modrinth)",
            takesValue: true
        }
    ],
    examples: [
        {
            description: "Add a mod by Modrinth URL",
            usage: "minepack add https://modrinth.com/mod/iris"
        },
        {
            description: "Add a mod by Modrinth ID",
            usage: "minepack add P7dR8mSH"
        },
        {
            description: "Add a mod by search term",
            usage: "minepack add sodium"
        },
        {
            description: "Add a mod by direct URL and specify all data",
            usage: "minepack add --url https://cdn.example.com/mod.jar --name MyMod --filename mod.jar --hash abc123 --hash-format sha1"
        },
        {
            description: "Add a mod and download the jar",
            usage: "minepack add sodium --download"
        }
    ],
    async execute(args, flags) {
        const modsDir = path.resolve(process.cwd(), "mods");
        if (!fs.existsSync(modsDir)) fs.mkdirSync(modsDir, { recursive: true });
        // Find pack.json in the parent directory
        const packJsonPath = path.resolve(process.cwd(), "pack.json");
        if (!fs.existsSync(packJsonPath)) {
            console.log(chalk.red("No pack.json found in the current directory. Please run this command from your pack root."));
            return;
        }
        const packJson = JSON.parse(fs.readFileSync(packJsonPath, "utf-8"));
        const packMeta = packJson; // Assume pack.json matches PackMeta
        const packGameVersion = packMeta.gameversion;
        const packModloader = packMeta.modloader?.name;
        console.log(chalk.gray(`[info] Detected pack version: ${packGameVersion}, modloader: ${packModloader}`));

        let modInput = args.mod as string;
        let modData: Partial<ModData> = {};
        let modrinthProject: any = null;
        let modrinthVersion: any = null;
        let cdnUrl: string | null = null;
        let hash: string | null = null;
        let hashFormat: string | null = null;
        let filename: string | null = null;
        let name: string | null = null;
        let update: any = undefined;
        let side: any = flags.side || ModSide.Both;

        // If --url is provided, treat as custom mod
        if (flags.url) {
            const url = flags.url as string;
            modData.name = (flags.name as string) || (args.mod as string);
            modData.filename = (flags.filename as string) || path.basename(url);
            modData.download = {
                url: url,
                'hash-format': (flags["hash-format"] as string) || "sha1",
                hash: (flags.hash as string) || ""
            };
            modData.side = side;
            console.log(chalk.gray(`[info] Adding custom mod: ${modData.name}`));
            if (flags.download) {
                // Download the file
                console.log(chalk.gray(`[info] Downloading mod from custom URL: ${url}`));
                const res = await fetch(url);
                if (!res.ok) throw new Error("Failed to download mod file");
                const filePath = path.join(modsDir, modData.filename ?? "mod.jar");
                const fileStream = fs.createWriteStream(filePath);
                if (!res.body) throw new Error("Response body is null");
                await new Promise<void>(async (resolve, reject) => {
                    if (res.body) {
                        const { Readable } = await import("stream");
                        Readable.fromWeb(res.body as any).pipe(fileStream);
                        fileStream.on("error", reject);
                        fileStream.on("finish", resolve);
                    } else {
                        reject(new Error("Response body is null"));
                    }
                });
                console.log(chalk.green(`Downloaded mod to ${filePath}`));
                return;
            } else {
                // Write .json stub
                const stubPath = path.join(modsDir, (modData.name ?? "mod") + ".json");
                fs.writeFileSync(stubPath, JSON.stringify(modData, null, 4));
                console.log(chalk.green(`Created mod stub at ${stubPath}`));
                return;
            }
        }

        // Try to parse as Modrinth URL
        let modrinthId = null;
        const modrinthUrlMatch = modInput.match(/modrinth\.com\/mod\/([\w-]+)/);
        if (modrinthUrlMatch) {
            modrinthId = modrinthUrlMatch[1];
        } else if (/^[a-zA-Z0-9]{5,}$/.test(modInput)) {
            modrinthId = modInput;
        }

        if (modrinthId) {
            modrinthProject = await getModrinthProject(modrinthId);
            if (!modrinthProject) {
                console.log(chalk.red("Could not find Modrinth project with that ID."));
                return;
            }
            console.log(chalk.gray(`[info] Found Modrinth project: ${modrinthProject.title} (${modrinthProject.id})`));
        } else {
            // Search Modrinth
            console.log(chalk.gray(`[info] Searching Modrinth for: ${modInput}`));
            const results = await searchModrinth(modInput);
            if (!results.length) {
                console.log(chalk.red("No mods found for that search term."));
                return;
            }
            console.log(chalk.bold("Select a mod:"));
            results.forEach((r: any, i: number) => {
                console.log(`  [${i + 1}] ${r.title} (${r.project_id}) - ${r.description}`);
            });
            const readline = await import('readline/promises');
            const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
            let idx = parseInt(await rl.question('Select mod [number]: '), 10) - 1;
            while (isNaN(idx) || idx < 0 || idx >= results.length) {
                idx = parseInt(await rl.question('Invalid selection. Select mod [number]: '), 10) - 1;
            }
            modrinthProject = await getModrinthProject(results[idx].project_id);
            await rl.close();
            console.log(chalk.gray(`[info] Selected Modrinth project: ${modrinthProject.title} (${modrinthProject.id})`));
        }

        // Get latest version for this mod, matching pack version and modloader
        console.log(chalk.gray(`[info] Fetching versions for mod: ${modrinthProject.title}`));
        modrinthVersion = await getModrinthVersion(modrinthProject.id, packGameVersion, packModloader);
        if (!modrinthVersion) {
            console.log(chalk.red(`Could not find a version for this mod matching Minecraft ${packGameVersion} and loader ${packModloader}.`));
            return;
        }
        cdnUrl = await getModrinthFileCdnUrl(modrinthVersion);
        if (!cdnUrl) {
            console.log(chalk.red("Could not find a downloadable file for this mod."));
            return;
        }
        filename = modrinthVersion.files[0].filename ?? "mod.jar";
        name = modrinthProject.title;
        hash = modrinthVersion.files[0].hashes.sha1 || modrinthVersion.files[0].hashes.sha512 || modrinthVersion.files[0].hashes.sha256 || "";
        hashFormat = hash
            ? (hash.length === 40 ? "sha1" : (hash.length === 128 ? "sha512" : "sha256"))
            : "";
        update = { 'mod-id': modrinthProject.id, version: modrinthVersion.version_number };

        modData = {
            name: name || "Unknown Mod",
            filename: filename as string,
            side,
            download: {
                url: cdnUrl,
                'hash-format': hashFormat,
                hash: hash || ""
            },
            update
        };

        if (flags.download) {
            // Download the file
            console.log(chalk.gray(`[info] Downloading mod from Modrinth CDN: ${cdnUrl}`));
            const res = await fetch(cdnUrl);
            const filePath = path.join(modsDir, filename ?? "mod.jar");
            const fileStream = fs.createWriteStream(filePath);
            if (!res.body) throw new Error("Response body is null");
            await new Promise<void>(async (resolve, reject) => {
                if (!res.body) {
                    reject(new Error("Response body is null"));
                    return;
                }
                const { Readable } = await import("stream");
                Readable.fromWeb(res.body as any).pipe(fileStream);
                fileStream.on("error", reject);
                fileStream.on("finish", resolve);
            });
            console.log(chalk.green(`Downloaded mod to ${filePath}`));
            return;
        } else {
            // Write .json stub (use mod name only)
            const stubPath = path.join(modsDir, (name || "mod") + ".json");
            fs.writeFileSync(stubPath, JSON.stringify(modData, null, 4));
            console.log(chalk.green(`Created mod stub at ${stubPath}`));
            return;
        }
    }
});

registerCommand(addCommand);

export { addCommand };
