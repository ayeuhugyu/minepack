import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { ModSide, type ModData } from "../lib/mod";

function getModsDir() {
    const modsDir = path.resolve(process.cwd(), "mods");
    if (!fs.existsSync(modsDir)) {
        throw new Error("No mods directory found in this pack.");
    }
    return modsDir;
}

function readAllMods(): ModData[] {
    const modsDir = getModsDir();
    const files = fs.readdirSync(modsDir).filter(f => f.endsWith(".json"));
    return files.map(f => {
        const data = JSON.parse(fs.readFileSync(path.join(modsDir, f), "utf-8"));
        return { ...data, _filename: f };
    });
}

const listCommand = new Command({
    name: "list",
    description: "List all mods in the pack.",
    arguments: [],
    flags: [
        { name: "url", aliases: [], description: "Show download URLs", takesValue: false },
        { name: "filename", aliases: [], description: "Show mod jar filenames", takesValue: false },
        { name: "side", aliases: [], description: "Show only mods for a specific side (client/server/both)", takesValue: true }
    ],
    examples: [
        { description: "List all mods", usage: "minepack list" },
        { description: "List all mods with URLs", usage: "minepack list --url" },
        { description: "List only client-side mods", usage: "minepack list --side client" }
    ],
    execute(_args, flags) {
        let mods = readAllMods();
        if (flags.side) {
            mods = mods.filter(m => (m.side || "both").toLowerCase() === String(flags.side).toLowerCase());
        }
        if (!mods.length) {
            console.log(chalk.yellow("No mods found in this pack."));
            return;
        }
        for (const mod of mods) {
            let line = chalk.green(mod.name || mod.filename || "?");
            if (flags.filename) line += chalk.gray(` [${mod.filename || "?"}]`);
            if (flags.url) line += chalk.cyan(` <${mod.download?.url || "?"}>`);
            if (!flags.side) line += chalk.magenta(` (${mod.side || "both"})`);
            console.log(line);
        }
    }
});

registerCommand(listCommand);

export { listCommand };
