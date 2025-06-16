import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import prettyBytes from "pretty-bytes";

registerCommand({
    name: "pack",
    aliases: ["showpack", "packinfo"],
    description: "Display metadata about the current minepack project, including total mod file size.",
    options: [],
    flags: [
        {
            name: "verbose",
            description: "Enable verbose output.",
            short: "v",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack pack",
        "minepack pack --verbose"
    ],
    execute: async ({ flags }) => {
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const stubs = pack.getStubs(!!flags.verbose);
        const totalSize = stubs.reduce((sum, stub) => sum + (stub.download?.size || 0), 0);
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
        console.log(chalk.greenBright.bold(pack.name) + chalk.gray(` by `) + chalk.cyanBright(pack.author));
        console.log(chalk.gray("Description: ") + chalk.whiteBright(pack.description));
        console.log(chalk.gray("Game Version: ") + chalk.yellowBright(pack.gameVersion));
        console.log(chalk.gray("Modloader: ") + chalk.yellowBright(`${pack.modloader.name} ${pack.modloader.version}`));
        console.log(chalk.gray("Number of Mods: ") + chalk.blueBright(stubs.length.toString()));
        console.log(chalk.gray("Total Mod File Size: ") + chalk.magentaBright(prettyBytes(totalSize)));
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
    }
});