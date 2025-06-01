import { commands, globalFlags, getCommand } from './command';
import './commands/help'; // Register help command

function parseArgs(argv: string[]) {
    const args: Record<string, string | boolean> = {};
    const flags: Record<string, boolean> = {};
    let commandName = '';
    let commandArgs: string[] = [];

    for (let i = 0; i < argv.length; i++) {
        const arg = argv[i];
        if (arg.startsWith('--')) {
            const flag = arg.slice(2);
            flags[flag] = true;
        } else if (arg.startsWith('-')) {
            const chars = arg.slice(1).split('');
            for (const c of chars) flags[c] = true;
        } else if (!commandName) {
            commandName = arg;
        } else {
            commandArgs.push(arg);
        }
    }
    return { commandName, commandArgs, flags };
}

function run() {
    const argv = process.argv.slice(2);
    const { commandName, commandArgs, flags } = parseArgs(argv);

    // If no command, show help
    if (!commandName) {
        commands['help'].execute({}, {});
        return;
    }

    // Check for global help flag
    if (flags['help'] || flags['h']) {
        commands['help'].execute({ command: commandName }, {});
        return;
    }

    const cmd = getCommand(commandName);
    if (!cmd) {
        console.error(`Unknown command: ${commandName}`);
        commands['help'].execute({}, {});
        return;
    }

    // Parse command arguments
    const argDefs = cmd.arguments || [];
    const parsedArgs: Record<string, string | boolean> = {};
    let argIndex = 0;
    for (const def of argDefs) {
        if (commandArgs[argIndex]) {
            parsedArgs[def.name] = commandArgs[argIndex];
            argIndex++;
        }
    }

    cmd.execute(parsedArgs, flags);
}

run();
