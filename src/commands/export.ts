import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { ContentType } from "../lib/mod";
import { execSync } from "child_process";
import { STUB_EXT, getContentFolders, getStubFilesFromTracked } from "../lib/packUtils";

const SUPPORTED_FOLDERS = getContentFolders();

async function collectContent() {
    const mods: any[] = [];
    for (const folder of SUPPORTED_FOLDERS) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith(STUB_EXT)) {
                const data = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf-8'));
                mods.push({ ...data, _folder: folder, _filename: file, _fullpath: path.join(dir, file) });
            }
        }
    }
    return mods;
}

async function collectOverrides(exportDir: string, modJsonFiles: Set<string>) {
    // Copy everything except .mp.json stubs into overrides
    const overridesDir = path.join(exportDir, "overrides");
    fs.mkdirSync(overridesDir, { recursive: true });
    for (const folder of SUPPORTED_FOLDERS) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith(STUB_EXT) || file.endsWith('.mrpack')) continue; // skip stubs and .mrpack files
            const full = path.join(dir, file);
            // Copy everything else
            const destFolder = path.join(overridesDir, folder);
            fs.mkdirSync(destFolder, { recursive: true });
            fs.copySync(full, path.join(destFolder, file));
            console.log(chalk.gray(`[info] Copied to overrides: ${folder}/${file}`));
        }
    }
    // Copy any other folders/files except .mp.json stubs, .mrpack files, and pack.mp.json
    for (const entry of fs.readdirSync(process.cwd())) {
        if (SUPPORTED_FOLDERS.includes(entry) || entry === "pack.mp.json" || entry === ".export" || entry.endsWith('.mrpack')) continue;
        const full = path.join(process.cwd(), entry);
        fs.copySync(full, path.join(overridesDir, entry));
        console.log(chalk.gray(`[info] Copied to overrides: ${entry}`));
    }
}

const exportCommand = new Command({
    name: "export",
    description: "Export the modpack to a distributable format (currently only 'modrinth').",
    arguments: [
        { name: "format", aliases: [], description: "Export format (e.g. modrinth)", required: true }
    ],
    flags: [
        { name: "side", aliases: [], description: "Only include content for this side (client/server/both)", takesValue: true },
        { name: "download", aliases: ["d"], description: "Forcibly download all mods and include them in overrides", takesValue: false }
    ],
    examples: [
        { description: "Export to Modrinth format", usage: "minepack export modrinth" },
        { description: "Export only client-side mods", usage: "minepack export modrinth --side client" },
        { description: "Export and forcibly download all mods", usage: "minepack export modrinth --download" }
    ],
    async execute(args, flags) {
        const format = args.format as string;
        if (format !== "modrinth") {
            console.log(chalk.red("Only 'modrinth' export is currently supported."));
            return;
        }
        const exportDir = path.resolve(process.cwd(), ".export");
        if (fs.existsSync(exportDir)) fs.rmSync(exportDir, { recursive: true, force: true });
        fs.mkdirSync(exportDir);
        console.log(chalk.gray(`[info] Created export working directory at ${exportDir}`));

        // Read pack.mp.json for meta
        const packJsonPath = path.resolve(process.cwd(), "pack.mp.json");
        if (!fs.existsSync(packJsonPath)) {
            console.log(chalk.red("No pack.mp.json found in the current directory."));
            return;
        }
        const packJson = JSON.parse(fs.readFileSync(packJsonPath, "utf-8"));

        // Collect all mod/content stubs
        // const side = flags.side as string | undefined;
        // const mods = await collectContent(side);
        // const modJsonFiles = new Set(mods.map(m => m._fullpath));
        // console.log(chalk.gray(`[info] Found ${mods.length} mod/content stubs for export.`));
        const side = flags.side as string | undefined;
        const mods = await collectContent();
        const modJsonFiles = new Set(mods.map(m => m._fullpath));
        console.log(chalk.gray(`[info] Found ${mods.length} mod/content stubs for export.`));

        // Build modrinth.index.json
        const index: any = {
            formatVersion: 1,
            game: "minecraft",
            versionId: packJson.version,
            name: packJson.name,
            summary: packJson.description || packJson.name,
            files: [],
            dependencies: {},
        };
        if (packJson.author) index.author = packJson.author;
        if (packJson.gameversion) index.dependencies.minecraft = packJson.gameversion;
        if (packJson.modloader) {
            // Map loader names for Modrinth compatibility
            let loaderName = packJson.modloader.name;
            if (loaderName === "fabric") loaderName = "fabric-loader";
            if (loaderName === "quilt") loaderName = "quilt-loader";
            index.dependencies[loaderName] = packJson.modloader.version;
        }

        for (const mod of mods) {
            // If forcibly download, download the file and put in overrides, and set envOnly: true
            if (flags.download && mod.download && mod.download.url) {
                const overridesModPath = path.join(exportDir, "overrides", mod._folder, mod.filename);
                fs.mkdirSync(path.dirname(overridesModPath), { recursive: true });
                console.log(chalk.gray(`[info] Downloading ${mod.name || mod.filename} to overrides...`));
                const res = await fetch(mod.download.url);
                if (!res.ok) {
                    console.log(chalk.red(`[warn] Failed to download ${mod.download.url}`));
                    continue;
                }
                const fileStream = fs.createWriteStream(overridesModPath);
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
                const fileSize = fs.statSync(overridesModPath).size;
                console.log(chalk.green(`[info] Downloaded and added to overrides: ${overridesModPath} (${fileSize} bytes)`));
                index.files.push({
                    path: `${mod._folder}/${mod.filename}`,
                    hashes: {
                        ...(mod.download.sha1 ? { sha1: mod.download.sha1 } : {}),
                        ...(mod.download.sha256 ? { sha256: mod.download.sha256 } : {}),
                        ...(mod.download.hash && mod.download["hash-format"] && !["sha1","sha256"].includes(mod.download["hash-format"]) ? { [mod.download["hash-format"]]: mod.download.hash } : {})
                    },
                    env: { client: "required", server: "required" },
                    downloads: [], // forcibly downloaded, so no download url
                    fileSize
                });
            } else {
                // Normal stub export
                let fileSize = mod.fileSize;
                // If not present, try to stat the file
                if (!fileSize && mod._folder && mod.filename) {
                    const filePath = path.resolve(process.cwd(), mod._folder, mod.filename);
                    if (fs.existsSync(filePath)) {
                        fileSize = fs.statSync(filePath).size;
                    }
                }
                if (!fileSize) fileSize = 0;
                // Determine env values based on --side flag
                let envClient = "required";
                let envServer = "required";
                if (flags.side === "client") {
                    envClient = mod.side === "client" || mod.side === "both" ? "required" : "unsupported";
                    envServer = mod.side === "server" ? "required" : "unsupported";
                } else if (flags.side === "server") {
                    envClient = mod.side === "server" ? "required" : "unsupported";
                    envServer = mod.side === "server" || mod.side === "both" ? "required" : "unsupported";
                }
                console.log(chalk.gray(`[info] Adding to index: ${mod._folder}/${mod.filename} (${fileSize} bytes, client: ${envClient}, server: ${envServer})`));
                index.files.push({
                    path: `${mod._folder}/${mod.filename}`,
                    hashes: {
                        ...(mod.download.sha1 ? { sha1: mod.download.sha1 } : {}),
                        ...(mod.download.sha256 ? { sha256: mod.download.sha256 } : {}),
                        ...(mod.download.hash && mod.download["hash-format"] && !["sha1","sha256"].includes(mod.download["hash-format"]) ? { [mod.download["hash-format"]]: mod.download.hash } : {})
                    },
                    env: {
                        client: envClient,
                        server: envServer
                    },
                    downloads: mod.download.url ? [mod.download.url] : [],
                    fileSize
                });
            }
        }
        // Write modrinth.index.json
        const indexPath = path.join(exportDir, "modrinth.index.json");
        fs.writeFileSync(indexPath, JSON.stringify(index, null, 4));
        console.log(chalk.green(`[info] Wrote modrinth.index.json with ${index.files.length} files.`));

        // Copy overrides
        await collectOverrides(exportDir, modJsonFiles);
        console.log(chalk.green(`[info] Copied overrides folder.`));

        // Zip and rename
        const zipPath = path.resolve(process.cwd(), `${packJson.name.replace(/\s+/g, "_")}-${packJson.version}.mrpack`);
        let zipped = false;
        let zipError = null;
        // Cross-platform zipping
        const isWin = process.platform === "win32";
        try {
            if (isWin) {
                // Try native PowerShell first
                const zipCmd = `Compress-Archive -Path '${exportDir}/*' -DestinationPath '${zipPath}.zip' -Force`;
                console.log(chalk.gray(`[info] Zipping export folder with native PowerShell...`));
                execSync(zipCmd, { stdio: "inherit", shell: "pwsh.exe" });
            } else {
                // Use system zip on Linux/macOS
                const zipCmd = `cd '${exportDir}' && zip -r '${zipPath}.zip' .`;
                console.log(chalk.gray(`[info] Zipping export folder with system zip...`));
                execSync(zipCmd, { stdio: "inherit", shell: "/bin/bash" });
            }
            zipped = true;
        } catch (err) {
            zipError = err;
        }
        if (!zipped) {
            console.log(chalk.red(`[error] Failed to zip export folder. Please zip the .export folder manually. Error: ${zipError}`));
            return;
        }
        fs.renameSync(`${zipPath}.zip`, zipPath);
        console.log(chalk.green(`[info] Exported modpack to ${zipPath}`));
        // Clean up
        fs.rmSync(exportDir, { recursive: true, force: true });
        console.log(chalk.gray(`[info] Cleaned up export working directory.`));
    }
});

registerCommand(exportCommand);

export { exportCommand };
