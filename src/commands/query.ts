import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { findMod, type ModData } from "../lib/mod";

function getModsDir() {
    const modsDir = path.resolve(process.cwd(), "mods");
    if (!fs.existsSync(modsDir)) {
        throw new Error("No mods directory found in this pack.");
    }
    return modsDir;
}

function readAllMods(): (ModData & { _filename?: string })[] {
    const modsDir = getModsDir();
    const files = fs.readdirSync(modsDir).filter(f => f.endsWith(".json"));
    return files.map(f => {
        const data = JSON.parse(fs.readFileSync(path.join(modsDir, f), "utf-8"));
        return { ...data, _filename: f };
    });
}

const queryCommand = new Command({
    name: "query",
    description: "Query if a mod exists in the modpack.",
    arguments: [
        { name: "mod", aliases: [], description: "The mod to search for (name, filename, or url)", required: true }
    ],
    flags: [],
    examples: [
        { description: "Query for a mod by name", usage: "minepack query sodium" },
        { description: "Query for a mod by filename", usage: "minepack query sodium-fabric-0.5.13+mc1.20.1.jar" },
        { description: "Query for a mod by url", usage: "minepack query https://cdn.modrinth.com/data/AANobbMI/versions/OihdIimA/sodium-fabric-0.5.13%2Bmc1.20.1.jar" }
    ],
    async execute(args) {
        const mods = readAllMods();
        const userInput = args.mod as string;
        const result = findMod(mods, userInput);
        if (result.mod) {
            console.log(chalk.green(`Found mod: ${result.mod.name || result.mod._filename}`));
            console.log(chalk.gray(JSON.stringify(result.mod, null, 4)));
        } else if (result.fuzzy && result.matches.length) {
            console.log(chalk.yellow("No exact match found. Top 5 fuzzy matches:"));
            result.matches.forEach((m, i) => {
                console.log(`  [${i + 1}] ${m.name || m._filename}`);
            });
            const readline = await import('readline/promises');
            const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
            let idx = parseInt(await rl.question('Select mod [number, or 0 to cancel]: '), 10) - 1;
            if (idx >= 0 && idx < result.matches.length) {
                const mod = result.matches[idx];
                console.log(chalk.green(`Selected mod: ${mod.name || mod._filename}`));
                console.log(chalk.gray(JSON.stringify(mod, null, 4)));
            } else {
                console.log(chalk.gray("No mod selected."));
            }
            await rl.close();
        } else {
            console.log(chalk.red("No mod found matching that input."));
        }
    }
});

registerCommand(queryCommand);

export { queryCommand };
