import { Command, registerCommand } from "../lib/command";
import chalk from "chalk";
import { VERSION } from "../version";

const versionCommand = new Command({
    name: "version",
    description: "Show the current minepack version.",
    arguments: [],
    flags: [],
    async execute() {
        console.log(chalk.green(`v${VERSION}`));
    }
});

registerCommand(versionCommand);

export { versionCommand };
