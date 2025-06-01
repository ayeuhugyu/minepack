import chalk from "chalk";
import {
    Command,
    registerCommand,
    commands,
    globalFlags,
    getCommand
} from '../command';

// Help command implementation
const helpCommand = new Command({
    name: 'help',
    description: 'Show help for a command or list all commands.',
    arguments: [
        {
            name: 'command',
            aliases: [],
            description: 'The command to show help for',
            required: false
        }
    ],
    examples: [
        {
            description: 'Show all commands',
            usage: 'minepack help'
        },
        {
            description: 'Show help for a specific command',
            usage: 'minepack help build'
        }
    ],
    execute(args) {
        const cmdName = args.command as string | undefined;
        if (!cmdName) {
            // List all commands
            console.log(chalk.bold("Available commands:"));
            for (const cmd of Object.values(commands)) {
                console.log(`  ${chalk.green(cmd.name)} - ${cmd.description}`);
            }
            console.log(`\nRun ${chalk.cyan("minepack help <command>")} to get more info about a command.`);
            return;
        }
        const cmd = getCommand(cmdName);
        if (!cmd) {
            console.error(chalk.red(`Unknown command: ${cmdName}`));
            return;
        }
        // Show manpage for the command
        console.log(chalk.bold("NAME"));
        console.log(`  ${chalk.green(cmd.name)} - ${cmd.description}\n`);
        if (cmd.arguments.length) {
            console.log(chalk.bold("ARGUMENTS:"));
            for (const arg of cmd.arguments) {
                const aliasStr = arg.aliases && arg.aliases.length ? chalk.gray(` (aliases: ${arg.aliases.join(", ")})`) : '';
                console.log(`  ${chalk.yellow(arg.name)}${aliasStr} - ${arg.description}`);
            }
            console.log("");
        }
        if (cmd.examples.length) {
            console.log(chalk.bold("EXAMPLES:"));
            for (const ex of cmd.examples) {
                console.log(`  ${chalk.gray("# " + ex.description)}`);
                console.log(`  ${chalk.cyan("$ " + ex.usage)}`);
            }
            console.log("");
        }
    }
});

registerCommand(helpCommand);

// Export for main CLI
export { helpCommand };