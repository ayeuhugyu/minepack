import { ContentType, ModSide, Content } from "../lib/mod";
import { Command, registerCommand } from "../lib/command";
import toml from "toml";
import chalk from "chalk";
import path from "path";
import fs from "fs-extra";

// For v4.1 upgrade
const MODRINTH_API = "https://api.modrinth.com/v2";
async function getModrinthProject(idOrSlug: string) {
    const url = `${MODRINTH_API}/project/${idOrSlug}`;
    const res = await fetch(url);
    if (!res.ok) return null;
    return await res.json();
}

async function importPackwiz(inputDir: string, outputDir: string) {
    // Find pack.toml (packwiz root)
    const cwd = inputDir;
    const packTomlPath = path.join(cwd, "pack.toml");
    if (!fs.existsSync(packTomlPath)) {
        console.log(chalk.red("No pack.toml found in the input directory. Please run this command from your packwiz pack root or specify the correct input directory."));
        return;
    }
    // Parse pack.toml
    const packToml = toml.parse(fs.readFileSync(packTomlPath, "utf-8"));
    // Build pack.mp.json
    const packJson: any = {
        name: packToml.name,
        author: packToml.author,
        version: packToml.version,
        gameversion: packToml.versions?.minecraft,
        modloader: packToml.versions?.fabric ? { name: "fabric", version: packToml.versions.fabric } : undefined
    };
    // Write pack.mp.json
    fs.writeFileSync(path.join(outputDir, "pack.mp.json"), JSON.stringify(packJson, null, 4));
    console.log(chalk.green(`Created pack.mp.json from pack.toml in ${outputDir}`));

    // Import content from all supported folders (mods, resourcepacks, shaderpacks, datapacks, plugins)
    const contentFolders = ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
    for (const folder of contentFolders) {
        const inDir = path.join(cwd, folder);
        const outDir = path.join(outputDir, folder);
        if (!fs.existsSync(inDir)) continue;
        fs.mkdirSync(outDir, { recursive: true });
        const files = fs.readdirSync(inDir).filter(f => f.endsWith(".pw.toml"));
        for (const file of files) {
            const modToml = toml.parse(fs.readFileSync(path.join(inDir, file), "utf-8"));
            // Use mod-id from TOML if available
            let modrinthId = modToml.update?.modrinth?.['mod-id'] || modToml['mod-id'] || modToml['modid'] || modToml['id'] || modToml.name;
            // Map folder to ContentType enum value
            let typeEnum = ContentType.Unknown;
            switch (folder) {
                case 'mods': typeEnum = ContentType.Mod; break;
                case 'resourcepacks': typeEnum = ContentType.Resourcepack; break;
                case 'shaderpacks': typeEnum = ContentType.Shaderpack; break;
                case 'datapacks': typeEnum = ContentType.Datapack; break;
                case 'plugins': typeEnum = ContentType.Plugin; break;
                default: typeEnum = ContentType.Unknown;
            }
            const modData = new Content({
                type: typeEnum,
                name: modToml.name,
                filename: modToml.filename,
                side: modToml.side || ModSide.Both,
                download: {
                    url: modToml.download?.url
                },
                hashes: {
                    sha1: modToml.download?.hash || "",
                    sha256: modToml.download?.sha256 || ""
                },
                update: modToml.update?.modrinth ? {
                    'mod-id': modrinthId,
                    version: modToml.update.modrinth.version
                } : undefined,
                dependencies: modToml.dependencies || [],
                fileSize: 0 // not known
            });
            // Sanitize name for filename
            let safeName = (modData.name || modData.filename || "content").replace(/[/\\?%*:|"<>.]+/g, '_').replace(/\.+$/, '').replace(/_+/g, '_');
            if (!safeName) safeName = "content";
            const stubPath = path.join(outDir, safeName + ".json");
            fs.writeFileSync(stubPath, JSON.stringify(modData, null, 4));
            console.log(chalk.green(`[import] Converted ${folder}: ${modData.name}`));
        }
        // Copy all non-.pw.toml files as overrides, preserving relative folder structure
        for (const file of fs.readdirSync(inDir)) {
            if (!file.endsWith(".pw.toml")) {
                const srcPath = path.join(inDir, file);
                const destPath = path.join(outputDir, folder, file);
                fs.copyFileSync(srcPath, destPath);
                console.log(chalk.gray(`[import] Copied override: ${file} -> ${destPath}`));
            }
        }
    }
    // Copy every file/folder from the packwiz project to the output, except for .pw.toml files (which are converted)
    const skipExt = ".pw.toml";
    const skipFolders = new Set(["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"]);
    function copyRecursive(srcDir: string, destDir: string) {
        for (const entry of fs.readdirSync(srcDir)) {
            const srcPath = path.join(srcDir, entry);
            const destPath = path.join(destDir, entry);
            const stat = fs.statSync(srcPath);
            if (stat.isDirectory()) {
                // Don't copy content folders here (handled above)
                if (skipFolders.has(entry)) continue;
                fs.mkdirSync(destPath, { recursive: true });
                copyRecursive(srcPath, destPath);
            } else {
                // Don't copy .pw.toml files (handled above)
                if (entry.endsWith(skipExt)) continue;
                fs.mkdirSync(path.dirname(destPath), { recursive: true });
                fs.copyFileSync(srcPath, destPath);
                console.log(chalk.gray(`[import] Copied file: ${srcPath} -> ${destPath}`));
            }
        }
    }
    copyRecursive(cwd, outputDir);
    console.log(chalk.bold.green("Packwiz import complete!"));
}

async function importOldMinepack(inputDir: string, outputDir: string) {
    // 1. Read old pack.json
    const oldPackPath = path.join(inputDir, "pack.json");
    if (!fs.existsSync(oldPackPath)) {
        console.log(chalk.red("No pack.json found in the input directory."));
        return;
    }
    const oldPack = JSON.parse(fs.readFileSync(oldPackPath, "utf-8"));
    // 2. Write new pack.mp.json
    fs.writeFileSync(path.join(outputDir, "pack.mp.json"), JSON.stringify(oldPack, null, 4));
    console.log(chalk.green(`Created pack.mp.json from pack.json in ${outputDir}`));
    // 3. Find all stubs (all .json files in content folders)
    const contentFolders = ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
    let tracked = [];
    for (const folder of contentFolders) {
        const inDir = path.join(inputDir, folder);
        const outDir = path.join(outputDir, folder);
        if (!fs.existsSync(inDir)) continue;
        fs.mkdirSync(outDir, { recursive: true });
        for (const file of fs.readdirSync(inDir)) {
            if (!file.endsWith(".json")) continue;
            const oldStubPath = path.join(inDir, file);
            let stubData;
            try {
                stubData = JSON.parse(fs.readFileSync(oldStubPath, "utf-8"));
            } catch {
                console.log(chalk.red(`[skip] Invalid JSON: ${oldStubPath}`));
                continue;
            }
            // Optionally validate stubData here (basic check)
            if (!stubData || typeof stubData !== 'object' || !stubData.name) {
                console.log(chalk.red(`[skip] Invalid stub: ${oldStubPath}`));
                continue;
            }
            // Write new stub with .mp.json extension
            let safeName = (stubData.name || stubData.filename || "content").replace(/[/\\?%*:|"<>.]+/g, '_').replace(/\.+$/, '').replace(/_+/g, '_');
            if (!safeName) safeName = "content";
            const newStubPath = path.join(outDir, safeName + ".mp.json");
            fs.writeFileSync(newStubPath, JSON.stringify(stubData, null, 4));
            tracked.push(path.relative(outputDir, newStubPath));
            console.log(chalk.green(`[import-old] Converted stub: ${file} -> ${safeName}.mp.json`));
        }
    }
    // 4. Write tracked.mp.json
    fs.writeFileSync(path.join(outputDir, "tracked.mp.json"), JSON.stringify(tracked, null, 2));
    console.log(chalk.green(`Created tracked.mp.json with ${tracked.length} stubs.`));
}

const importCommand = new Command({
    name: "import",
    description: "Import a modpack from another format (currently only packwiz is supported, or v4.1 for legacy upgrade).",
    arguments: [
        { name: "format", aliases: [], description: "Format to import from (e.g. packwiz, old, v4.1)", required: true }
    ],
    flags: [
        { name: "input", aliases: ["i"], description: "Input directory (packwiz project root or old minepack root)", takesValue: true },
        { name: "output", aliases: ["o"], description: "Output directory (minepack project root)", takesValue: true },
        { name: "old", aliases: [], description: "Import from old Minepack pack.json format", takesValue: false }
    ],
    examples: [
        { description: "Import a packwiz pack", usage: "minepack import packwiz --input ./packwizpack --output ./minepackproject" },
        { description: "Import an old Minepack project", usage: "minepack import old --input ./oldminepack --output ./minepackproject" },
        { description: "Upgrade legacy v4.1 stubs to new format", usage: "minepack import v4.1 --input ./oldminepack --output ./minepackproject" }
    ],
    async execute(args, flags) {
        if (args.format === "old") {
            const inputDir = typeof flags.input === 'string' ? path.resolve(flags.input) : process.cwd();
            const outputDir = typeof flags.output === 'string' ? path.resolve(flags.output) : process.cwd();
            await importOldMinepack(inputDir, outputDir);
            return;
        }
        if (args.format === "packwiz") {
            const inputDir = typeof flags.input === 'string' ? path.resolve(flags.input) : process.cwd();
            const outputDir = typeof flags.output === 'string' ? path.resolve(flags.output) : process.cwd();
            await importPackwiz(inputDir, outputDir);
            return;
        }
        if (args.format === "v4.1") {
            // Upgrade legacy v4.1 stubs to new format
            const inputDir = typeof flags.input === 'string' ? path.resolve(flags.input) : process.cwd();
            const outputDir = typeof flags.output === 'string' ? path.resolve(flags.output) : process.cwd();
            const folders = ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
            let tracked = [];
            for (const folder of folders) {
                const inDir = path.join(inputDir, folder);
                const outDir = path.join(outputDir, folder);
                if (!fs.existsSync(inDir)) continue;
                fs.mkdirSync(outDir, { recursive: true });
                for (const file of fs.readdirSync(inDir)) {
                    if (!file.endsWith(".json")) continue;
                    const oldStubPath = path.join(inDir, file);
                    let stubData;
                    try { stubData = JSON.parse(fs.readFileSync(oldStubPath, "utf-8")); } catch { continue; }
                    if (!stubData || typeof stubData !== 'object' || !stubData.name) continue;
                    // Re-fetch from Modrinth if possible to get hashes
                    const regex = /https:\/\/cdn.modrinth.com\/data\/(.+?)\/versions\/(.+?)\//
                    let modrinthId = stubData.update?.['mod-id'] || regex.exec(stubData.download?.url)?.[1] || stubData.name;
                    let modrinthVersionId = stubData.update?.version || regex.exec(stubData.download?.url)?.[2] || "";
                    let newStub = { ...stubData };
                    let hashes = { sha1: "", sha256: "" };
                    if (stubData.update?.['mod-id']) {
                        // Try to fetch hashes from Modrinth
                        try {
                            const modrinthProject = await getModrinthProject(modrinthId);
                            if (modrinthProject) {
                                const versionRes = await fetch(`${MODRINTH_API}/project/${modrinthProject.id}/version`);
                                if (versionRes.ok) {
                                    const versions = await versionRes.json();
                                    const found = versions.find((v: any) => v.version_number === modrinthVersionId);
                                    if (found && found.files && found.files[0]) {
                                        hashes.sha1 = found.files[0].hashes.sha1 || "";
                                        hashes.sha256 = found.files[0].hashes.sha256 || "";
                                    }
                                    // const versions = await versionRes.json();
                                    // const found = versions.find((v: any) => v.version_number === stubData.update.version);
                                    // if (found && found.files && found.files[0]) {
                                    //     hashes.sha1 = found.files[0].hashes.sha1 || "";
                                    //     hashes.sha256 = found.files[0].hashes.sha256 || "";
                                    // }
                                }
                            }
                        } catch {}
                    }
                    newStub.hashes = hashes;
                    // Remove old hash-format/hash fields if present
                    if (newStub.download) {
                        delete newStub.download['hash-format'];
                        delete newStub.download.hash;
                    }
                    // Write new stub with .mp.json extension
                    let safeName = (newStub.name || newStub.filename || "content").replace(/[/\\?%*:|"<>.]+/g, '_').replace(/\.+$/, '').replace(/_+/g, '_');
                    if (!safeName) safeName = "content";
                    const stubPath = path.join(outDir, safeName + ".mp.json");
                    fs.writeFileSync(stubPath, JSON.stringify(newStub, null, 4));
                    tracked.push(path.join(folder, safeName + ".mp.json"));
                    console.log(chalk.green(`[import] Upgraded ${folder}: ${newStub.name}`));
                }
            }
            // Write tracked.mp.json
            fs.writeFileSync(path.join(outputDir, "tracked.mp.json"), JSON.stringify(tracked, null, 2));
            console.log(chalk.green(`Created tracked.mp.json with ${tracked.length} upgraded stubs.`));
            console.log(chalk.bold.green("Legacy v4.1 upgrade complete!"));
            return;
        }
        console.log(chalk.red("Only 'packwiz', 'old' and 'v4.1' import formats are currently supported."));
    }
});

registerCommand(importCommand);

export { importCommand };
