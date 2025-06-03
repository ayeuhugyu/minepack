import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import toml from "toml";
import { Content, ContentType, ModSide, HashFormat } from "../lib/mod";

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
                    url: modToml.download?.url,
                    'hash-format': modToml.download?.['hash-format'] || HashFormat.Sha1,
                    hash: modToml.download?.hash || ""
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

const importCommand = new Command({
    name: "import",
    description: "Import a modpack from another format (currently only packwiz is supported).",
    arguments: [
        { name: "format", aliases: [], description: "Format to import from (e.g. packwiz)", required: true }
    ],
    flags: [
        { name: "input", aliases: ["i"], description: "Input directory (packwiz project root)", takesValue: true },
        { name: "output", aliases: ["o"], description: "Output directory (minepack project root)", takesValue: true }
    ],
    examples: [
        { description: "Import a packwiz pack", usage: "minepack import packwiz --input ./packwizpack --output ./minepackproject" }
    ],
    async execute(args, flags) {
        if (args.format !== "packwiz") {
            console.log(chalk.red("Only 'packwiz' import is currently supported."));
            return;
        }
        const inputDir = typeof flags.input === 'string' ? path.resolve(flags.input) : process.cwd();
        const outputDir = typeof flags.output === 'string' ? path.resolve(flags.output) : process.cwd();
        await importPackwiz(inputDir, outputDir);
    }
});

registerCommand(importCommand);

export { importCommand };
