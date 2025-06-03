import path from "path";
import fs from "fs-extra";
import chalk from "chalk";
import { Command, registerCommand } from "../lib/command";
import { ModSide, type ModData } from "../lib/mod";
import { STUB_EXT, getContentFolders, getStubFilesFromTracked } from "../lib/packUtils";

function getModsDir() {
    const modsDir = path.resolve(process.cwd(), "mods");
    if (!fs.existsSync(modsDir)) {
        throw new Error("No mods directory found in this pack.");
    }
    return modsDir;
}

function readAllMods(): ModData[] {
    const modsDir = getModsDir();
    const files = fs.readdirSync(modsDir).filter(f => f.endsWith(STUB_EXT));
    return files.map(f => {
        const data = JSON.parse(fs.readFileSync(path.join(modsDir, f), "utf-8"));
        return { ...data, _filename: f };
    });
}

function readAllContent() {
    // Use tracked.mp.json for stubs
    const rootDir = process.cwd();
    const stubFiles = getStubFilesFromTracked(rootDir);
    let all: Array<{ type: string, name: string, filename: string, side?: string, download?: any, fileSize?: number, _folder: string, _isStub: boolean }> = [];
    for (const stubRelPath of stubFiles) {
        const absPath = path.join(rootDir, stubRelPath);
        if (!fs.existsSync(absPath)) continue;
        try {
            const data = JSON.parse(fs.readFileSync(absPath, "utf-8"));
            const folder = stubRelPath.split(path.sep)[0];
            all.push({ ...data, _folder: folder, _isStub: true, filename: data.filename || path.basename(stubRelPath) });
        } catch {}
    }
    const folders = getContentFolders();
    for (const folder of folders) {
        const dir = path.resolve(process.cwd(), folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            const ext = path.extname(file).toLowerCase();
            if ([".jar", ".zip", ".mcpack", ".datapack", ".litemod"].includes(ext)) {
                all.push({
                    type: folder.slice(-1) === "s" ? folder.slice(0, -1) : folder, // crude type guess
                    name: file,
                    filename: file,
                    _folder: folder,
                    _isStub: false
                });
            }
        }
    }
    return all;
}

const listCommand = new Command({
    name: "list",
    description: "List all mods and content in the pack (mods, resourcepacks, shaderpacks, etc).",
    arguments: [],
    flags: [
        { name: "url", aliases: ["u"], description: "Show download URLs", takesValue: false },
        { name: "filename", aliases: ["f"], description: "Show file names", takesValue: false },
        { name: "side", aliases: ["s"], description: "Show only content for a specific side (client/server/both)", takesValue: true },
        { name: "type", aliases: ["t"], description: "Show only a specific content type (mod/resourcepack/shaderpack/etc)", takesValue: true },
        { name: "clean", aliases: ["c"], description: "Show only the mod/content name (no extra info)", takesValue: false }
    ],
    examples: [
        { description: "List all content", usage: "minepack list" },
        { description: "List all with URLs", usage: "minepack list --url" },
        { description: "List only client-side mods", usage: "minepack list --side client --type mod" }
    ],
    execute(_args, flags) {
        let content = readAllContent();
        if (flags.side) {
            content = content.filter(m => (m.side || "both").toLowerCase() === String(flags.side).toLowerCase());
        }
        if (flags.type) {
            content = content.filter(m => (m.type || m._folder).toLowerCase() === String(flags.type).toLowerCase());
        }
        if (!content.length) {
            console.log(chalk.yellow("No content found in this pack."));
            return;
        }
        content.forEach((item, idx) => {
            if (flags.clean) {
                console.log(item.name || item.filename || "?");
                return;
            }
            let line = chalk.yellow(`[${idx + 1}] `) + chalk.green(item.name || item.filename || "?");
            line += chalk.gray(` [${item._folder}]`);
            if (flags.filename) line += chalk.gray(` [${item.filename || "?"}]`);
            if (flags.url && item.download?.url) line += chalk.cyan(` <${item.download.url}>`);
            if (!flags.side && item.side) line += chalk.magenta(` (${item.side})`);
            if (!item._isStub) line += chalk.yellow(" [file]");
            console.log(line);
        });
    }
});

registerCommand(listCommand);

export { listCommand };
