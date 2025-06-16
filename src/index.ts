import { commands, registerCommand } from "./lib/command";
import type { CommandFlag, CommandContext, CommandDefinition } from "./lib/command";
import chalk from "chalk";

// Import commands; this needs to be done since the files won't run unless they're imported by something
import "./commands/version";
import "./commands/help";
import "./commands/test";
import "./commands/init";
import "./commands/search";
import "./commands/add";
import "./commands/remove";
import "./commands/list";
import "./commands/map";
import "./commands/update";
import "./commands/query";

// Command line parser
function parseArgs<Flags extends readonly CommandFlag[]>(argv: string[], command: { flags: Flags }): CommandContext<Flags> {
    const flags: any = {};
    const options: string[] = [];
    const flagDefs = new Map<string, any>();
    const shortFlagDefs = new Map<string, any>();
    for (const flag of command.flags) {
        flagDefs.set(`--${flag.name}`, flag);
        if (flag.short) shortFlagDefs.set(flag.short, flag);
    }
    let i = 0;
    while (i < argv.length) {
        const arg = argv[i];
        if (flagDefs.has(arg)) {
            // Long flag
            const flagDef = flagDefs.get(arg);
            if (flagDef.takesValue) {
                const value = argv[i + 1];
                if (!value || value.startsWith("-")) throw new Error(`Flag ${arg} requires a value.`);
                flags[flagDef.name] = value;
                i += 2;
            } else {
                flags[flagDef.name] = true;
                i++;
            }
        } else if (arg.startsWith("-") && !arg.startsWith("--")) {
            // Short flag(s), e.g. -fg or -f -g
            const shorts = arg.slice(1);
            let consumed = false;
            for (let j = 0; j < shorts.length; j++) {
                const short = shorts[j];
                const flagDef = shortFlagDefs.get(short);
                if (!flagDef) throw new Error(`Unknown flag: -${short}`);
                if (flagDef.takesValue) {
                    // Value can be attached (-svalue) or next arg (-s value)
                    let value: string | undefined;
                    if (j < shorts.length - 1) {
                        value = shorts.slice(j + 1);
                        flags[flagDef.name] = value;
                        consumed = true;
                        break;
                    } else {
                        value = argv[i + 1];
                        if (!value || value.startsWith("-")) throw new Error(`Flag -${short} requires a value.`);
                        flags[flagDef.name] = value;
                        i++;
                    }
                } else {
                    flags[flagDef.name] = true;
                }
            }
            i += consumed ? 1 : 1;
        } else if (arg.startsWith("-")) {
            throw new Error(`Unknown flag: ${arg}`);
        } else {
            options.push(arg);
            i++;
        }
    }
    return { flags, options };
}

// Main entry
function main() {
    const [, , cmdName, ...args] = process.argv;
    const command = commands.find(
        c => c.name === cmdName || (c.aliases && c.aliases.includes(cmdName))
    );
    if (!command) {
        console.error(chalk.redBright.bold(` ✖  Unknown command: ${cmdName}; use \`minepack help\` to see available commands.`));
        process.exit(1);
    }
    try {
        const ctx = parseArgs(args, command);
        command.execute(ctx);
    } catch (err) {
        console.error(chalk.redBright.bold((err as Error).message));
        process.exit(1);
    }
}

main();