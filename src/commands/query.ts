import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { findMod, type ModData } from "../lib/mod";

function getContentFolders() {
    return ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
}

function getModsDir() {
    const modsDir = path.resolve(process.cwd(), "mods");
    if (!fs.existsSync(modsDir)) {
        throw new Error("No mods directory found in this pack.");
    }
    return modsDir;
}

function readAllContent(): (ModData & { _filename?: string, _folder?: string })[] {
    const folders = getContentFolders();
    let all: Array<ModData & { _filename?: string, _folder?: string }> = [];
    for (const folder of folders) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith(".json")) {
                const data = JSON.parse(fs.readFileSync(path.join(dir, file), "utf-8"));
                all.push({ ...data, _filename: file, _folder: folder });
            }
        }
    }
    return all;
}

const queryCommand = new Command({
    name: "query",
    description: "Query if a mod or content exists in the modpack.",
    arguments: [
        { name: "mod", aliases: [], description: "The mod/content to search for (name, filename, or url)", required: true }
    ],
    flags: [],
    examples: [
        { description: "Query for a mod by name", usage: "minepack query sodium" },
        { description: "Query for a resourcepack by name", usage: "minepack query MyResourcepack" },
        { description: "Query for a mod by filename", usage: "minepack query sodium-fabric-0.5.13+mc1.20.1.jar" },
        { description: "Query for a mod by url", usage: "minepack query https://cdn.modrinth.com/data/AANobbMI/versions/OihdIimA/sodium-fabric-0.5.13%2Bmc1.20.1.jar" }
    ],
    async execute(args) {
        const content = readAllContent();
        const userInput = args.mod as string;
        console.log(chalk.gray(`[info] Loaded ${content.length} content stubs from all folders`));
        // First search: only stubs (JSON files)
        let result = findMod(content, userInput);
        let found = null;
        let searchStage = 'stubs';
        if (result.mod) {
            console.log(chalk.gray(`[info] Exact match found in stubs: ${result.mod.name || result.mod._filename} [${result.mod._folder}]`));
            found = result.mod;
        } else if (result.fuzzy && result.matches.length) {
            console.log(chalk.yellow("No exact match found in stubs. Top 5 fuzzy matches:"));
            result.matches.forEach((m, i) => {
                console.log(`  [${i + 1}] ${m.name || m._filename} [${m._folder}]`);
            });
            const readline = await import('readline/promises');
            const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
            let idx = parseInt(await rl.question('Select content to query [number, or 0 to cancel]: '), 10) - 1;
            if (idx >= 0 && idx < result.matches.length) {
                found = result.matches[idx];
                console.log(chalk.gray(`[info] User selected: ${found.name || found._filename} [${found._folder}]`));
            } else {
                console.log(chalk.gray("No content selected."));
            }
            await rl.close();
        }
        // If not found in stubs, search for .jar files by filename
        if (!found) {
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
                found = exact;
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
                    let idx = parseInt(await rl.question('Select .jar to query [number, or 0 to cancel]: '), 10) - 1;
                    if (idx >= 0 && idx < matches.length) {
                        found = matches[idx];
                        console.log(chalk.gray(`[info] User selected: ${found._filename} [${found._folder}]`));
                        searchStage = 'jar';
                    } else {
                        console.log(chalk.gray("No .jar selected."));
                    }
                    await rl.close();
                }
            }
        }
        if (found && found._filename && found._folder) {
            console.log(chalk.green(`[query] ${found._filename} [${found._folder}] (${searchStage === 'jar' ? '.jar file' : 'stub'})`));
            if (searchStage === 'stubs') {
                // Print all metadata for stub
                console.log(JSON.stringify(found, null, 2));
            } else {
                // .jar file: print basic info
                const dir = path.resolve(process.cwd(), found._folder);
                const filePath = path.join(dir, found._filename);
                const stats = fs.statSync(filePath);
                console.log(`Filename: ${found._filename}`);
                console.log(`Folder: ${found._folder}`);
                console.log(`Size: ${stats.size} bytes`);
                console.log(`Path: ${filePath}`);
            }
        } else if (!found) {
            console.log(chalk.red("No content found matching that input (in stubs or .jar files)."));
        }
    }
});

registerCommand(queryCommand);

export { queryCommand };
