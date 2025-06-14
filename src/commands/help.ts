import { commands, registerCommand, type CommandFlag } from "../lib/command";
import chalk from "chalk";

registerCommand({
    name: "help",
    aliases: [],
    description: "Displays a list of commands or details for a specific command.",
    options: [
        {
            name: "command",
            description: "The command to get help for.",
            required: false,
            exampleValues: ["version"],
        }
    ],
    flags: [],
    exampleUsage: [
        "minepack help",
        "minepack help version"
    ],
    execute: async ({ flags, options }) => {
        const commandName = options[0];

        if (commandName) {
            const command = commands.find(c => c.name === commandName || (c.aliases && c.aliases.includes(commandName)));
            if (!command) {
                console.error(chalk.redBright.bold(" ✖ Unknown command: ") + chalk.whiteBright(` ${commandName} `) + chalk.gray("; use ") + chalk.blueBright("minepack help") + chalk.gray(" to see available commands."));
                return;
            }

            console.log(chalk.gray('────────────────────────────────────────────────────────────'));
            console.log(
                chalk.blueBright.bold('minepack ') +
                chalk.bold(command.name) +
                ((command.aliases.length > 0) ? chalk.gray(` [aliases: ${command.aliases.map(a => chalk.yellowBright(a)).join(chalk.gray(", "))}]`) : "")
            );
            console.log(chalk.whiteBright(command.description));
            if (command.options.length > 0) {
                console.log("\n" + chalk.greenBright.bold("⚙ OPTIONS"));
                let optCount = 0;
                command.options.forEach(opt => {
                    const isLastOpt = ++optCount === command.options.length;
                    console.log(
                        `  ${chalk.bold.dim.greenBright(isLastOpt ? "╰╴" : "├╴")}${chalk.greenBright(opt.name)}${opt.required ? chalk.redBright(' (required)') : chalk.gray(' [optional]')}: ${chalk.whiteBright(opt.description)}`
                    );
                    if (opt.exampleValues) {
                        console.log(chalk.gray(`      examples:`));
                        opt.exampleValues.forEach(val => {
                            console.log(`        ${chalk.magentaBright('→')} minepack ${chalk.bold.white(command.name)} ${chalk.blueBright(val)}`);
                        });
                    }
                });
            }

            if (command.flags.length > 0) {
                console.log("\n" + chalk.yellowBright.bold("⚐ FLAGS"));
                let count = 0;
                command.flags.forEach((flag: CommandFlag) => {
                    const isLastFlag = ++count === command.flags.length;
                    console.log(
                        `  ${chalk.bold.dim.yellowBright(isLastFlag ? "╰╴" : "├╴")}${chalk.yellowBright(`--${flag.name}`)}${flag.short ? ` ${chalk.gray(`(-${flag.short})`)}` : ""}${flag.takesValue ? chalk.gray(" <value>") : ""}: ${chalk.whiteBright(flag.description)}`
                    );
                    if (flag.exampleValues) {
                        console.log(chalk.gray(`      examples:`));
                        flag.exampleValues.forEach(val => {
                            console.log(`        ${chalk.magentaBright('→')} minepack ${chalk.bold(command.name)} ${flag.short ? chalk.gray(`-${flag.short}`) : chalk.yellowBright(`--${flag.name}`)} ${chalk.blueBright(val)}`);
                        });
                    }
                });
            }

            if (command.exampleUsage.length > 0) {
                console.log("\n" + chalk.magentaBright.bold("🗋 EXAMPLE USAGE"));
                const flagRegex = /(--[\w-]+|-[a-zA-Z]{1,})/g;
                command.exampleUsage.forEach((example, i) => {
                    const isLastEx = i === command.exampleUsage.length - 1;
                    const coloredExample = example.replace(flagRegex, match => chalk.gray(match));
                    console.log(`  ${chalk.bold.dim.magentaBright(isLastEx ? "╰╴" : "├╴")}${coloredExample}`);
                });
            }
            console.log(chalk.gray('────────────────────────────────────────────────────────────'));
        } else {
            console.log(chalk.gray('────────────────────────────────────────────────────────────'));
            console.log(chalk.magentaBright.bold("🕮  AVAILABLE COMMANDS"));
            commands.forEach((command, i) => {
                const isLastCmd = i === commands.length - 1;
                console.log(`  ${chalk.bold.dim.blueBright(isLastCmd ? "╰╴" : "├╴")}${chalk.blueBright(command.name)}: ${chalk.whiteBright(command.description)}`);
            });
            console.log("\n" + chalk.gray("Use ") + chalk.blueBright("minepack help <command>") + chalk.gray(" to get more details on a specific command."));
            console.log(chalk.gray('────────────────────────────────────────────────────────────'));
        }
    }
});