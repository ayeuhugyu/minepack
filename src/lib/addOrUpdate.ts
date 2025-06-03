import chalk from "chalk";
import path from "path";
import fs from "fs-extra";
import { ModSide, ContentType, modrinthTypeToContentType, contentTypeToFolder } from "./mod";
import { STUB_EXT, addStubToTracked } from "./packUtils";

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

async function getModrinthFileCdnUrl(version: any) {
    if (!version || !version.files || !version.files.length) return null;
    // Prefer primary file
    const file = version.files.find((f: any) => f.primary) || version.files[0];
    return file.url;
}

/**
 * Core logic for adding or updating a mod/content from Modrinth or direct URL.
 * Used by both add and update commands.
 * @param {object} opts - Options for the operation
 * @param {string} opts.input - Modrinth URL, ID, search term, or direct URL
 * @param {object} opts.flags - CLI flags (download, side, url, name, filename, hash, hash-format, type)
 * @param {object} opts.packMeta - The pack.mp.json metadata
 * @param {boolean} [opts.interactive=true] - Whether to prompt user for ambiguous cases
 * @param {function} [opts.onPrompt] - Optional callback for user prompts (for update command)
 * @returns {Promise<object>} - Result object with status, messages, and file info
 */
export async function addOrUpdateContent({ input, flags, packMeta, interactive = true, onPrompt, verbose = false }: any) {
    let modInput = input;
    let modData: any = {};
    let modrinthProject: any = null;
    let modrinthVersion: any = null;
    let cdnUrl: string | null = null;
    let hash: string | null = null;
    let hashFormat: string | null = null;
    let filename: string | null = null;
    let name: string | null = null;
    let update: any = undefined;
    let side: any = flags.side || ModSide.Both;
    let contentType: ContentType = ContentType.Mod; // default
    let folder: string = "mods";
    let outDir: string;
    const packGameVersion = packMeta.gameversion;
    const packModloader = packMeta.modloader?.name;

    // Helper to sanitize names for filenames
    function sanitize(name: string) {
        return (name || "content").replace(/[/\\?%*:|"<>.]+/g, '_').replace(/\.+$/, '').replace(/_+/g, '_');
    }

    // If --url is provided, or input is a direct non-Modrinth URL, just download the file to mods/ (unless --type is set)
    if ((flags.url || (modInput.startsWith('https://') && !modInput.includes('modrinth.com/mod/')))) {
        const url = (flags.url as string) || modInput;
        if (flags.type) {
            contentType = flags.type as ContentType;
            folder = contentTypeToFolder(contentType);
        } else {
            folder = 'mods';
        }
        outDir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(outDir)) {
            if (verbose) console.log(chalk.gray(`[info] Creating ${folder} directory at ${outDir}`));
            fs.mkdirSync(outDir, { recursive: true });
        }
        if (!flags.name && !flags.filename && !flags["hash-format"] && !flags.hash && !flags.download && !flags.type) {
            const filename = path.basename(url.split('?')[0]);
            const filePath = path.join(outDir, filename);
            if (verbose) console.log(chalk.gray(`[info] Downloading file from URL: ${url}`));
            const res = await fetch(url);
            if (!res.ok) throw new Error("Failed to download file");
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
            let stats = { size: 0 };
            try { stats = fs.statSync(filePath); } catch {}
            const stubPath = path.join(outDir, sanitize(filename) + STUB_EXT);
            if (verbose) console.log(chalk.gray(`[info] Writing stub to ${stubPath}`));
            fs.writeFileSync(stubPath, JSON.stringify({
                type: contentType,
                name: filename,
                filename,
                side,
                download: { url, 'hash-format': '', hash: '' },
                fileSize: stats.size || 0
            }, null, 4));
            addStubToTracked(process.cwd(), path.relative(process.cwd(), stubPath));
            if (verbose) console.log(chalk.green(`[info] Downloaded file to ${filePath}`));
            return { status: 'success', message: `Downloaded file to ${filePath}` };
        }
        // Otherwise, use the stub logic (for advanced/manual use)
        modData.type = contentType;
        modData.name = (flags.name as string) || modInput;
        modData.filename = (flags.filename as string) || path.basename(url);
        modData.download = {
            url: url,
            'hash-format': (flags["hash-format"] as string) || "sha1",
            hash: (flags.hash as string) || ""
        };
        modData.side = side;
        const maybeFile = path.join(outDir, modData.filename);
        if (fs.existsSync(maybeFile)) {
            modData.fileSize = fs.statSync(maybeFile).size;
        } else {
            modData.fileSize = 0;
        }
        if (verbose) console.log(chalk.gray(`[info] Adding custom ${contentType}: ${modData.name}`));
        if (flags.download) {
            if (verbose) console.log(chalk.gray(`[info] Downloading file from custom URL: ${url}`));
            const res = await fetch(url);
            if (!res.ok) throw new Error("Failed to download file");
            const filePath = path.join(outDir, modData.filename ?? "file.bin");
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
            if (fs.existsSync(filePath)) {
                modData.fileSize = fs.statSync(filePath).size;
            } else {
                modData.fileSize = 0;
            }
            if (verbose) console.log(chalk.green(`[info] Downloaded file to ${filePath}`));
            return { status: 'success', message: `Downloaded file to ${filePath}` };
        } else {
            const stubPath = path.join(outDir, sanitize(modData.name ?? "content") + STUB_EXT);
            if (verbose) console.log(chalk.gray(`[info] Writing stub to ${stubPath}`));
            fs.writeFileSync(stubPath, JSON.stringify(modData, null, 4));
            addStubToTracked(process.cwd(), path.relative(process.cwd(), stubPath));
            if (verbose) console.log(chalk.green(`[info] Created stub at ${stubPath}`));
            return { status: 'success', message: `Created stub at ${stubPath}` };
        }
    }

    // Try to parse as Modrinth URL
    let modrinthId = modInput;
    const modrinthUrlMatch = modInput.match(/modrinth\.com\/mod\/([\w-]+)/);
    if (modrinthUrlMatch) {
        modrinthId = modrinthUrlMatch[1];
        if (verbose) console.log(chalk.gray(`[info] Parsed Modrinth ID from URL: ${modrinthId}`));
    } else if (/^[a-zA-Z0-9]{5,}$/.test(modInput)) {
        modrinthId = modInput;
        if (verbose) console.log(chalk.gray(`[info] Using Modrinth ID: ${modrinthId}`));
    }

    let triedSearch = false;
    let searchResults = [];
    let searchTerm = modInput;

    // Try to fetch by ID first
    if (modrinthId) {
        if (verbose) console.log(chalk.gray(`[info] Fetching Modrinth project for ID: ${modrinthId}`));
        modrinthProject = await getModrinthProject(modrinthId);
        if (!modrinthProject) {
            if (verbose) console.log(chalk.yellow(`[info] Could not find Modrinth project with that ID. Falling back to search...`));
            triedSearch = true;
            searchResults = await searchModrinth(modrinthId);
        }
    } else {
        // No ID, must search
        triedSearch = true;
        if (verbose) console.log(chalk.gray(`[info] Searching Modrinth for: ${searchTerm}`));
        searchResults = await searchModrinth(searchTerm);
    }

    if (triedSearch) {
        if (!searchResults.length) {
            if (verbose) console.log(chalk.red(`[info] No content found for that search term.`));
            return { status: 'notfound', message: 'No content found for that search term.' };
        }
        if (interactive) {
            // Prompt user to select
            const readline = await import('readline/promises');
            const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
            searchResults.forEach((r: any, i: number) => {
                console.log(`  [${i + 1}] ${r.title} (${r.project_id}) - ${r.project_type} - ${r.description}`);
            });
            let idx = parseInt(await rl.question('Select project [number]: '), 10) - 1;
            while (isNaN(idx) || idx < 0 || idx >= searchResults.length) {
                idx = parseInt(await rl.question('Invalid selection. Select project [number]: '), 10) - 1;
            }
            if (verbose) console.log(chalk.gray(`[info] Fetching Modrinth project for selection: ${searchResults[idx].project_id}`));
            modrinthId = searchResults[idx].project_id;
            await rl.close();
        } else if (onPrompt) {
            const idx = await onPrompt(searchResults);
            if (typeof idx !== 'number' || idx < 0 || idx >= searchResults.length) {
                if (verbose) console.log(chalk.yellow(`[info] Skipped by user.`));
                return { status: 'skipped', message: 'Skipped by user.' };
            }
            modrinthId = searchResults[idx].project_id;
        } else {
            if (verbose) console.log(chalk.yellow(`[info] No selection made.`));
            return { status: 'skipped', message: 'No selection made.' };
        }
        // Try fetching again with new modrinthId
        if (verbose) console.log(chalk.gray(`[info] Fetching Modrinth project for ID: ${modrinthId}`));
        modrinthProject = await getModrinthProject(modrinthId);
        if (!modrinthProject) {
            if (verbose) console.log(chalk.red(`[info] Could not find Modrinth project after search.`));
            return { status: 'notfound', message: 'Could not find Modrinth project with that ID or search.' };
        }
    }

    // Fetch Modrinth project
    if (verbose) console.log(chalk.gray(`[info] Fetching Modrinth project for ID: ${modrinthId}`));
    modrinthProject = await getModrinthProject(modrinthId);
    if (!modrinthProject) {
        if (verbose) console.log(chalk.red(`[info] Could not find Modrinth project with that ID.`));
        return { status: 'notfound', message: 'Could not find Modrinth project with that ID.' };
    }
    contentType = modrinthTypeToContentType(modrinthProject.project_type);
    folder = contentTypeToFolder(contentType);
    outDir = path.resolve(process.cwd(), folder);
    if (!fs.existsSync(outDir)) {
        if (verbose) console.log(chalk.gray(`[info] Creating ${folder} directory at ${outDir}`));
        fs.mkdirSync(outDir, { recursive: true });
    }

    // Get latest version for this project, matching if possible
    if (verbose) console.log(chalk.gray(`[info] Fetching versions for: ${modrinthProject.title}`));
    let versions = [];
    let versionMatch = null;
    let loaderMatch = null;
    let anyVersion = null;
    const versionRes = await fetch(`${MODRINTH_API}/project/${modrinthProject.id}/version`);
    if (versionRes.ok) {
        versions = await versionRes.json();
        if (contentType === ContentType.Mod) {
            versionMatch = versions.find((v: any) => v.game_versions.includes(packGameVersion) && v.loaders.includes(packModloader));
            loaderMatch = versions.find((v: any) => v.loaders.includes(packModloader));
            anyVersion = versions[0];
            if (!versionMatch && !loaderMatch && !anyVersion) {
                if (interactive) {
                    const readline = await import('readline/promises');
                    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
                    let answer = await rl.question(`No version found for ${modrinthProject.title} with Minecraft ${packGameVersion} and loader ${packModloader}. [i]gnore/[r]emove? `);
                    await rl.close();
                    if (answer.trim().toLowerCase().startsWith('r')) {
                        if (verbose) console.log(chalk.red(`[info] User chose to remove.`));
                        return { status: 'remove', message: 'User chose to remove.' };
                    } else {
                        if (verbose) console.log(chalk.yellow(`[info] User chose to ignore.`));
                        return { status: 'skipped', message: 'User chose to ignore.' };
                    }
                } else if (onPrompt) {
                    const action = await onPrompt([], modrinthProject);
                    if (action === 'remove') {
                        if (verbose) console.log(chalk.red(`[info] User chose to remove.`));
                        return { status: 'remove', message: 'User chose to remove.' };
                    }
                    if (verbose) console.log(chalk.yellow(`[info] User chose to ignore.`));
                    return { status: 'skipped', message: 'User chose to ignore.' };
                } else {
                    if (verbose) console.log(chalk.yellow(`[info] No suitable version.`));
                    return { status: 'skipped', message: 'No suitable version.' };
                }
            }
        } else {
            versionMatch = versions.find((v: any) => v.game_versions.includes(packGameVersion));
            anyVersion = versions[0];
        }
    }
    modrinthVersion = versionMatch || loaderMatch || anyVersion;
    if (!modrinthVersion) {
        if (verbose) console.log(chalk.red(`[info] Could not find a suitable version for this content.`));
        return { status: 'notfound', message: 'Could not find a suitable version for this content.' };
    }
    if (verbose) console.log(chalk.gray(`[info] Found version: ${modrinthVersion.version_number}`));
    cdnUrl = await getModrinthFileCdnUrl(modrinthVersion);
    if (!cdnUrl) {
        if (verbose) console.log(chalk.red(`[info] Could not find a downloadable file for this content.`));
        return { status: 'notfound', message: 'Could not find a downloadable file for this content.' };
    }
    filename = modrinthVersion.files[0].filename ?? "file.bin";
    name = modrinthProject.title;
    hash = modrinthVersion.files[0].hashes.sha1 || modrinthVersion.files[0].hashes.sha512 || modrinthVersion.files[0].hashes.sha256 || "";
    hashFormat = hash
        ? (hash.length === 40 ? "sha1" : (hash.length === 128 ? "sha512" : "sha256"))
        : "";
    update = { 'mod-id': modrinthProject.id, version: modrinthVersion.version_number };

    modData = {
        type: contentType,
        name: name || "Unknown Content",
        filename: filename as string,
        side,
        download: {
            url: cdnUrl,
            'hash-format': hashFormat,
            hash: hash || ""
        },
        update,
        fileSize: 0 // default
    };
    if (verbose) console.log(chalk.gray(`[info] Final mod data: ${JSON.stringify(modData, null, 2)}`));

    // Write stub file
    const stubPath = path.join(outDir, sanitize(modData.name) + STUB_EXT);
    if (verbose) console.log(chalk.gray(`[info] Writing stub to ${stubPath}`));
    fs.writeFileSync(stubPath, JSON.stringify(modData, null, 4));
    addStubToTracked(process.cwd(), path.relative(process.cwd(), stubPath));
    if (verbose) console.log(chalk.green(`[info] Created stub at ${stubPath}`));
    return { status: 'success', message: `Created stub at ${stubPath}` };
}
