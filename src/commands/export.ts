import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { ContentType } from "../lib/mod";
import { execSync } from "child_process";

const SUPPORTED_FOLDERS = [
    "mods",
    "resourcepacks",
    "shaderpacks",
    "datapacks",
    "plugins"
];

async function collectContent(sideFilter?: string) {
    const mods: any[] = [];
    for (const folder of SUPPORTED_FOLDERS) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith('.json')) {
                const data = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf-8'));
                if (!sideFilter || (data.side || "both") === sideFilter || data.side === "both") {
                    mods.push({ ...data, _folder: folder, _filename: file, _fullpath: path.join(dir, file) });
                }
            }
        }
    }
    return mods;
}

async function collectOverrides(exportDir: string, modJsonFiles: Set<string>) {
    // Copy everything except .json stubs into overrides
    const overridesDir = path.join(exportDir, "overrides");
    fs.mkdirSync(overridesDir, { recursive: true });
    for (const folder of SUPPORTED_FOLDERS) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith('.json') || file.endsWith('.mrpack')) continue; // skip stubs and .mrpack files
            const full = path.join(dir, file);
            // Copy everything else
            const destFolder = path.join(overridesDir, folder);
            fs.mkdirSync(destFolder, { recursive: true });
            fs.copySync(full, path.join(destFolder, file));
        }
    }
    // Copy any other folders/files except .json stubs, .mrpack files, and pack.json
    for (const entry of fs.readdirSync(process.cwd())) {
        if (SUPPORTED_FOLDERS.includes(entry) || entry === "pack.json" || entry === ".export" || entry.endsWith('.mrpack')) continue;
        const full = path.join(process.cwd(), entry);
        fs.copySync(full, path.join(overridesDir, entry));
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

        // Read pack.json for meta
        const packJsonPath = path.resolve(process.cwd(), "pack.json");
        if (!fs.existsSync(packJsonPath)) {
            console.log(chalk.red("No pack.json found in the current directory."));
            return;
        }
        const packMeta = JSON.parse(fs.readFileSync(packJsonPath, "utf-8"));

        // Collect all mod/content stubs
        const side = flags.side as string | undefined;
        const mods = await collectContent(side);
        const modJsonFiles = new Set(mods.map(m => m._fullpath));
        console.log(chalk.gray(`[info] Found ${mods.length} mod/content stubs for export.`));

        // Build modrinth.index.json
        const index: any = {
            formatVersion: 1,
            game: "minecraft",
            versionId: packMeta.version,
            name: packMeta.name,
            summary: packMeta.description || packMeta.name,
            files: [],
            dependencies: {},
        };
        if (packMeta.author) index.author = packMeta.author;
        if (packMeta.gameversion) index.dependencies.minecraft = packMeta.gameversion;
        if (packMeta.modloader) {
            // Map loader names for Modrinth compatibility
            let loaderName = packMeta.modloader.name;
            if (loaderName === "fabric") loaderName = "fabric-loader";
            if (loaderName === "quilt") loaderName = "quilt-loader";
            index.dependencies[loaderName] = packMeta.modloader.version;
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
                    hashes: mod.download.hash ? { [mod.download["hash-format"] || "sha1"]: mod.download.hash } : {},
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
                console.log(chalk.gray(`[info] Adding to index: ${mod._folder}/${mod.filename} (${fileSize} bytes)`));
                index.files.push({
                    path: `${mod._folder}/${mod.filename}`,
                    hashes: mod.download.hash ? { [mod.download["hash-format"] || "sha1"]: mod.download.hash } : {},
                    env: {
                        client: mod.side === "server" ? "unsupported" : "required",
                        server: mod.side === "client" ? "unsupported" : "required"
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
        const zipPath = path.resolve(process.cwd(), `${packMeta.name.replace(/\s+/g, "_")}-${packMeta.version}.mrpack`);
        let zipped = false;
        let zipError = null;
        // Try native PowerShell first
        try {
            const zipCmd = `Compress-Archive -Path '${exportDir}/*' -DestinationPath '${zipPath}.zip' -Force`;
            console.log(chalk.gray(`[info] Zipping export folder with native PowerShell...`));
            execSync(zipCmd, { stdio: "inherit", shell: "pwsh.exe" });
            zipped = true;
        } catch (err) {
            zipError = err;
            try {
                const zipCmd = `powershell Compress-Archive -Path '${exportDir}/*' -DestinationPath '${zipPath}.zip' -Force`;
                console.log(chalk.gray(`[info] Zipping export folder with 'powershell' command...`));
                execSync(zipCmd, { stdio: "inherit" });
                zipped = true;
            } catch (err2) {
                zipError = err2;
            }
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
