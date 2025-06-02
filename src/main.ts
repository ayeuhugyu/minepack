import { commands, globalFlags, getCommand } from './lib/command';
import './commands/help'; // Register help command
import './commands/version'; // Register version command
import './commands/init'; // Register init command
import './commands/add'; // Register add command
import './commands/list'; // Register list command
import './commands/query'; // Register remove command
import './commands/search'; // Register search command
import './commands/remove'; // Register remove command
import './commands/export'; // Register export command
import './commands/import'; // Register import command
import './commands/update'; // Register update command
import './commands/selfupdate'; // Register self-update command

function parseArgs(argv: string[], flagDefs: import('./lib/command').CommandFlag[] = []) {
    const args: Record<string, string | boolean> = {};
    const flags: Record<string, string | boolean> = {};
    let commandName = '';
    let commandArgs: string[] = [];

    for (let i = 0; i < argv.length; i++) {
        const arg = argv[i];
        if (arg.startsWith('--')) {
            let flag = arg.slice(2);
            let value: string | boolean = true;
            // Handle --no-flag for negatable flags
            if (flag.startsWith('no-')) {
                flag = flag.slice(3);
                value = false;
            }
            // Check if this flag takes a value
            const def = flagDefs.find(f => f.name === flag || (f.aliases && f.aliases.includes(flag)));
            if (def && def.takesValue && argv[i + 1] && !argv[i + 1].startsWith('-')) {
                value = argv[i + 1];
                i++;
            }
            flags[flag] = value;
        } else if (arg.startsWith('-') && arg.length > 1) {
            const chars = arg.slice(1).split('');
            for (let j = 0; j < chars.length; j++) {
                const c = chars[j];
                const def = flagDefs.find(f => f.aliases && f.aliases.includes(c));
                if (def && def.takesValue && argv[i + 1] && !argv[i + 1].startsWith('-')) {
                    flags[def.name] = argv[i + 1];
                    i++;
                    break;
                } else {
                    flags[c] = true;
                }
            }
        } else if (!commandName) {
            commandName = arg;
        } else {
            commandArgs.push(arg);
        }
    }
    return { commandName, commandArgs, flags };
}

function filterBooleanFlags(flags: Record<string, string | boolean>): Record<string, boolean> {
    const out: Record<string, boolean> = {};
    for (const [k, v] of Object.entries(flags)) {
        if (typeof v === 'boolean') out[k] = v;
    }
    return out;
}

function run() {
    const argv = process.argv.slice(2);
    // Get command name first to get flagDefs
    let { commandName } = parseArgs(argv);
    const cmd = getCommand(commandName);
    const flagDefs = cmd ? cmd.flags || [] : [];
    const { commandArgs, flags } = parseArgs(argv, flagDefs);

    // If no command, show help
    if (!commandName) {
        commands['help'].execute({}, {}, filterBooleanFlags(flags));
        return;
    }

    // Check for global help flag
    if (flags['help'] || flags['h']) {
        commands['help'].execute({ command: commandName }, {}, filterBooleanFlags(flags));
        return;
    }

    if (!cmd) {
        console.error(`Unknown command: ${commandName}`);
        commands['help'].execute({}, {}, filterBooleanFlags(flags));
        return;
    }

    // Parse command arguments and command-specific flags
    const argDefs = cmd.arguments || [];
    const parsedArgs: Record<string, string | boolean> = {};
    const parsedCmdFlags: Record<string, string | boolean> = {};
    let argIndex = 0;
    for (const def of argDefs) {
        if (commandArgs[argIndex]) {
            parsedArgs[def.name] = commandArgs[argIndex];
            argIndex++;
        }
    }
    // Separate command-specific flags from global flags
    for (const flagDef of flagDefs) {
        if (flags[flagDef.name] !== undefined) parsedCmdFlags[flagDef.name] = flags[flagDef.name];
        if (flagDef.aliases) {
            for (const alias of flagDef.aliases) {
                if (flags[alias] !== undefined) parsedCmdFlags[flagDef.name] = flags[alias];
            }
        }
    }
    // Remove command-specific flags from global flags
    const globalFlagsOnly: Record<string, string | boolean> = { ...flags };
    for (const key of Object.keys(parsedCmdFlags)) {
        delete globalFlagsOnly[key];
    }

    // Only pass string|boolean for command flags, but only boolean for global flags
    const filteredCmdFlags: Record<string, string | boolean> = {};
    for (const [k, v] of Object.entries(parsedCmdFlags)) {
        if (typeof v === 'string' || typeof v === 'boolean') filteredCmdFlags[k] = v;
    }
    cmd.execute(parsedArgs, filteredCmdFlags, filterBooleanFlags(globalFlagsOnly));
}

run();
