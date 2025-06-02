import chalk from "chalk";
import path from "path";
import fs from "fs-extra";
import { Command, registerCommand } from "../lib/command";
import { addOrUpdateContent } from "../lib/addOrUpdate";
import { contentTypeToFolder } from "../lib/mod";

const updateCommand = new Command({
    name: "update",
    description: "Update all mods and content in your pack to the latest version for the current Minecraft version and modloader. This may change mod versions and loader compatibility if your pack.json has changed. If a mod cannot be updated, you will be prompted to ignore or remove it.",
    async execute() {
        const packJsonPath = path.resolve(process.cwd(), "pack.json");
        if (!fs.existsSync(packJsonPath)) {
            console.log(chalk.red("No pack.json found in the current directory. Please run this command from your pack root."));
            return;
        }
        const packJson = JSON.parse(fs.readFileSync(packJsonPath, "utf-8"));
        // Find all .json stubs in all content folders
        const folders = ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
        let stubs: { file: string, data: any, folder: string }[] = [];
        for (const folder of folders) {
            const dir = path.resolve(process.cwd(), folder);
            if (!fs.existsSync(dir)) continue;
            for (const file of fs.readdirSync(dir)) {
                if (file.endsWith(".json")) {
                    try {
                        const data = JSON.parse(fs.readFileSync(path.join(dir, file), "utf-8"));
                        stubs.push({ file: path.join(dir, file), data, folder });
                    } catch {}
                }
            }
        }
        if (!stubs.length) {
            console.log(chalk.yellow("No mod/content stubs found to update."));
            return;
        }
        for (const stub of stubs) {
            const { data, file, folder } = stub;
            let input = data.update?.['mod-id'] || data.download?.url || data.name;
            let flags = { ...data, ...data.download, type: data.type };
            // Remove fields that shouldn't be flags
            delete flags.download;
            delete flags.update;
            delete flags.fileSize;
            delete flags.type;
            // Use addOrUpdateContent with non-interactive prompt
            const result = await addOrUpdateContent({
                input,
                flags: { ...flags, type: data.type },
                packMeta: packJson,
                interactive: false,
                onPrompt: async (results: any[], modrinthProject?: any) => {
                    if (results && results.length) {
                        // Pick first result automatically for update
                        return 0;
                    } else if (modrinthProject) {
                        // Prompt user to ignore or remove
                        const readline = await import('readline/promises');
                        const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
                        let answer = await rl.question(`No version found for ${modrinthProject.title} with current pack version/loader. [i]gnore/[r]emove? `);
                        await rl.close();
                        if (answer.trim().toLowerCase().startsWith('r')) return 'remove';
                        return 'ignore';
                    }
                    return 'ignore';
                }
            });
            if (result.status === 'remove') {
                fs.unlinkSync(file);
                console.log(chalk.red(`[removed] ${file}`));
            } else if (result.status === 'skipped') {
                console.log(chalk.yellow(`[skipped] ${file}`));
            } else if (result.status === 'success') {
                console.log(chalk.green(`[updated] ${file}`));
            } else if (result.status === 'notfound') {
                console.log(chalk.red(`[not found] ${file}`));
            } else {
                console.log(chalk.gray(`[info] ${file}: ${result.message}`));
            }
        }
        console.log(chalk.bold.green("Update complete."));
    }
});

registerCommand(updateCommand);

export { updateCommand };
