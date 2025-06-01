// Command system for minepack CLI
// Defines Command class, argument structure, and help system

export interface CommandArgument {
    name: string;
    aliases?: string[];
    description: string;
    required?: boolean;
}

export interface CommandExample {
    description: string;
    usage: string;
}

export interface CommandOptions {
    name: string;
    description: string;
    arguments?: CommandArgument[];
    examples?: CommandExample[];
    execute: (args: Record<string, string | boolean>, globalFlags: Record<string, boolean>) => Promise<void> | void;
}

export class Command {
    name: string;
    description: string;
    arguments: CommandArgument[];
    examples: CommandExample[];
    execute: (args: Record<string, string | boolean>, globalFlags: Record<string, boolean>) => Promise<void> | void;

    constructor(options: CommandOptions) {
        this.name = options.name;
        this.description = options.description;
        this.arguments = options.arguments || [];
        this.examples = options.examples || [];
        this.execute = options.execute;
    }
}

// Command registry and global flag system
export const commands: Record<string, Command> = {};
export const commandAliases: Record<string, string> = {};
export const globalFlags = [
    { name: 'help', aliases: ['h'], description: 'Show help for a command' }
];

export function registerCommand(cmd: Command, aliases: string[] = []) {
    commands[cmd.name] = cmd;
    for (const alias of aliases) {
        commandAliases[alias] = cmd.name;
    }
}

export function getCommand(name: string): Command | undefined {
    return commands[name] || commands[commandAliases[name]];
}
