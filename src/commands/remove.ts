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

const removeCommand = new Command({
    name: "remove",
    description: "Remove a mod from the modpack.",
    arguments: [
        { name: "mod", aliases: [], description: "The mod to remove (name, filename, or url)", required: true }
    ],
    flags: [],
    examples: [
        { description: "Remove a mod by name", usage: "minepack remove sodium" },
        { description: "Remove a mod by filename", usage: "minepack remove sodium-fabric-0.5.13+mc1.20.1.jar" },
        { description: "Remove a mod by url", usage: "minepack remove https://cdn.modrinth.com/data/AANobbMI/versions/OihdIimA/sodium-fabric-0.5.13%2Bmc1.20.1.jar" }
    ],
    async execute(args) {
        const mods = readAllMods();
        const userInput = args.mod as string;
        console.log(chalk.gray(`[info] Loaded ${mods.length} mods from mods directory`));
        const result = findMod(mods, userInput);
        let toRemove: (ModData & { _filename?: string }) | null = null;
        if (result.mod) {
            console.log(chalk.gray(`[info] Exact match found: ${result.mod.name || result.mod._filename}`));
            toRemove = result.mod;
        } else if (result.fuzzy && result.matches.length) {
            console.log(chalk.yellow("No exact match found. Top 5 fuzzy matches:"));
            result.matches.forEach((m, i) => {
                console.log(`  [${i + 1}] ${m.name || m._filename}`);
            });
            const readline = await import('readline/promises');
            const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
            let idx = parseInt(await rl.question('Select mod to remove [number, or 0 to cancel]: '), 10) - 1;
            if (idx >= 0 && idx < result.matches.length) {
                toRemove = result.matches[idx];
                console.log(chalk.gray(`[info] User selected: ${toRemove.name || toRemove._filename}`));
            } else {
                console.log(chalk.gray("No mod selected."));
            }
            await rl.close();
        }
        if (toRemove && toRemove._filename) {
            const modsDir = getModsDir();
            const filePath = path.join(modsDir, toRemove._filename);
            console.log(chalk.gray(`[info] Removing file: ${filePath}`));
            fs.unlinkSync(filePath);
            console.log(chalk.green(`Removed mod: ${toRemove.name || toRemove._filename}`));
        } else if (!toRemove) {
            console.log(chalk.red("No mod found matching that input."));
        }
    }
});

registerCommand(removeCommand);

export { removeCommand };
