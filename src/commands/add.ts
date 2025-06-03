import chalk from "chalk";
import path from "path";
import fs from "fs-extra";
import { Command, registerCommand } from "../lib/command";
import { addOrUpdateContent } from "../lib/addOrUpdate";

const addCommand = new Command({
    name: "add",
    description: "Add a mod to your modpack from Modrinth or a direct URL.",
    arguments: [
        {
            name: "mod",
            aliases: [],
            description: "Modrinth URL, mod ID, or search term.",
            required: true
        }
    ],
    flags: [
        {
            name: "download",
            aliases: ["D"],
            description: "Download the mod jar directly instead of creating a .json stub.",
            takesValue: false
        },
        {
            name: "side",
            aliases: ["s"],
            description: "Which side to use this mod on (client/server/both)",
            takesValue: true
        },
        {
            name: "url",
            aliases: [],
            description: "Direct download URL (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "name",
            aliases: [],
            description: "Name of the mod (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "filename",
            aliases: [],
            description: "Filename for the mod (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "hash",
            aliases: [],
            description: "Hash for the mod file (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "hash-format",
            aliases: [],
            description: "Hash format (sha1, sha256, etc) (if not using Modrinth)",
            takesValue: true
        },
        {
            name: "type",
            aliases: ["t"],
            description: "Content type (mod/resourcepack/datapack) (if not using Modrinth)",
            takesValue: true
        }
    ],
    examples: [
        {
            description: "Add a mod by Modrinth URL",
            usage: "minepack add https://modrinth.com/mod/iris"
        },
        {
            description: "Add a mod by Modrinth ID",
            usage: "minepack add P7dR8mSH"
        },
        {
            description: "Add a mod by search term",
            usage: "minepack add sodium"
        },
        {
            description: "Add a mod by direct URL and specify all data",
            usage: "minepack add --url https://cdn.example.com/mod.jar --name MyMod --filename mod.jar --hash abc123 --hash-format sha1"
        },
        {
            description: "Add a mod and download the jar",
            usage: "minepack add sodium --download"
        }
    ],
    async execute(args, flags) {
        const packJsonPath = path.resolve(process.cwd(), "pack.mp.json");
        if (!fs.existsSync(packJsonPath)) {
            console.log(chalk.red("No pack.mp.json found in the current directory. Please run this command from your pack root."));
            return;
        }
        const packJson = JSON.parse(fs.readFileSync(packJsonPath, "utf-8"));
        console.log(chalk.gray(`[info] Loaded pack.mp.json: version=${packJson.gameversion}, modloader=${packJson.modloader?.name}`));
        const result = await addOrUpdateContent({
            input: args.mod,
            flags,
            packMeta: packJson,
            interactive: true,
            verbose: true,
            sanitizeName: true
        });
        if (result && result.message) {
            if (result.status === 'success') {
                console.log(chalk.green(result.message));
            } else if (result.status === 'notfound') {
                console.log(chalk.red(result.message));
            } else {
                console.log(chalk.yellow(result.message));
            }
        }
    }
});

registerCommand(addCommand);

export { addCommand };
