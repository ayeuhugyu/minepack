import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { findMod, type ModData } from "../lib/mod";
import { removeStubFromTracked, STUB_EXT, getContentFolders, getStubFilesFromTracked } from "../lib/packUtils";

function getModsDir() {
    const modsDir = path.resolve(process.cwd(), "mods");
    if (!fs.existsSync(modsDir)) {
        throw new Error("No mods directory found in this pack.");
    }
    return modsDir;
}

function readAllContent(includeJars = false): (ModData & { _filename?: string, _folder?: string })[] {
    const folders = getContentFolders();
    let all: Array<ModData & { _filename?: string, _folder?: string }> = [];
    for (const folder of folders) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith(STUB_EXT)) {
                const data = JSON.parse(fs.readFileSync(path.join(dir, file), "utf-8"));
                all.push({ ...data, _filename: file, _folder: folder });
            } else if (includeJars && file.endsWith(".jar")) {
                // Guess ContentType from folder
                let typeEnum = "mod";
                switch (folder) {
                    case 'mods': typeEnum = "mod"; break;
                    case 'resourcepacks': typeEnum = "resourcepack"; break;
                    case 'shaderpacks': typeEnum = "shaderpack"; break;
                    case 'datapacks': typeEnum = "datapack"; break;
                    case 'plugins': typeEnum = "plugin"; break;
                    default: typeEnum = "unknown";
                }
                all.push({
                    _filename: file,
                    _folder: folder,
                    name: file,
                    filename: file,
                    type: typeEnum as any,
                    download: { url: '', 'hash-format': '', hash: '' },
                    fileSize: fs.statSync(path.join(dir, file)).size
                });
            }
        }
    }
    return all;
}

const removeCommand = new Command({
    name: "remove",
    description: "Remove a mod or content from the modpack.",
    arguments: [
        { name: "mod", aliases: [], description: "The mod/content to remove (name, filename, or url)", required: true }
    ],
    flags: [],
    examples: [
        { description: "Remove a mod by name", usage: "minepack remove sodium" },
        { description: "Remove a resourcepack by name", usage: "minepack remove MyResourcepack" },
        { description: "Remove a mod by filename", usage: "minepack remove sodium-fabric-0.5.13+mc1.20.1.jar" },
        { description: "Remove a mod by url", usage: "minepack remove https://cdn.modrinth.com/data/AANobbMI/versions/OihdIimA/sodium-fabric-0.5.13%2Bmc1.20.1.jar" }
    ],
    async execute(args) {
        const userInput = args.mod as string;
        const rootDir = process.cwd();
        const folders = getContentFolders();
        let toRemove = null;
        let searchStage = 'stubs';
        // Try exact stub file match first
        for (const folder of folders) {
            const stubPath = path.join(rootDir, folder, userInput + STUB_EXT);
            if (fs.existsSync(stubPath)) {
                try {
                    const data = JSON.parse(fs.readFileSync(stubPath, "utf-8"));
                    toRemove = { ...data, _filename: userInput + STUB_EXT, _folder: folder };
                    break;
                } catch {}
            }
        }
        let result = null;
        if (!toRemove) {
            const content = readAllContent();
            result = findMod(content, userInput);
            if (result.mod) {
                toRemove = result.mod;
            } else if (result.fuzzy && result.matches.length) {
                console.log(chalk.yellow("No exact match found in stubs. Top 5 fuzzy matches:"));
                result.matches.forEach((m, i) => {
                    console.log(`  [${i + 1}] ${m.name || m._filename} [${m._folder}]`);
                });
                const readline = await import('readline/promises');
                const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
                let idx = parseInt(await rl.question('Select content to remove [number, or 0 to cancel]: '), 10) - 1;
                if (idx >= 0 && idx < result.matches.length) {
                    toRemove = result.matches[idx];
                    console.log(chalk.gray(`[info] User selected: ${toRemove.name || toRemove._filename} [${toRemove._folder}]`));
                } else {
                    console.log(chalk.gray("No content selected."));
                }
                await rl.close();
            }
        }
        // If not found in stubs, search for .jar files by filename
        if (!toRemove) {
            console.log(chalk.gray("[info] No match found in stubs. Searching for .jar files by filename..."));
            const folders = getContentFolders();
            let jarCandidates = [];
            for (const folder of folders) {
                const dir = path.resolve(process.cwd(), folder);
                if (!fs.existsSync(dir)) continue;
                for (const file of fs.readdirSync(dir)) {
                    if (file.endsWith(".jar")) {
                        jarCandidates.push({ _filename: file, _folder: folder });
                    }
                }
            }
            // Try exact match by filename
            let exact = jarCandidates.find(j => j._filename === userInput);
            if (exact) {
                console.log(chalk.gray(`[info] Exact .jar filename match: ${exact._filename} [${exact._folder}]`));
                toRemove = exact;
                searchStage = 'jar';
            } else {
                // Fuzzy: substring match (case-insensitive)
                let matches = jarCandidates.filter(j => j._filename.toLowerCase().includes(userInput.toLowerCase()));
                if (matches.length) {
                    console.log(chalk.yellow("No exact .jar match. Top 5 fuzzy .jar matches:"));
                    matches.slice(0, 5).forEach((m, i) => {
                        console.log(`  [${i + 1}] ${m._filename} [${m._folder}]`);
                    });
                    const readline = await import('readline/promises');
                    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
                    let idx = parseInt(await rl.question('Select .jar to remove [number, or 0 to cancel]: '), 10) - 1;
                    if (idx >= 0 && idx < matches.length) {
                        toRemove = matches[idx];
                        console.log(chalk.gray(`[info] User selected: ${toRemove._filename} [${toRemove._folder}]`));
                        searchStage = 'jar';
                    } else {
                        console.log(chalk.gray("No .jar selected."));
                    }
                    await rl.close();
                }
            }
        }
        if (toRemove && toRemove._filename && toRemove._folder) {
            const dir = path.resolve(process.cwd(), toRemove._folder);
            const filePath = path.join(dir, toRemove._filename);
            fs.unlinkSync(filePath);
            if (toRemove._filename.endsWith(STUB_EXT)) {
                removeStubFromTracked(process.cwd(), path.relative(process.cwd(), filePath));
            }
            console.log(chalk.green(`Removed: ${toRemove._filename} [${toRemove._folder}] (${searchStage === 'jar' ? '.jar file' : 'stub'})`));
        } else if (!toRemove) {
            console.log(chalk.red("No content found matching that input (in stubs or .jar files)."));
        }
    }
});

registerCommand(removeCommand);

export { removeCommand };
