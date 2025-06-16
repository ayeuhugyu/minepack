// Command definition and base class for command structure
export interface CommandOption {
    name: string;
    description: string;
    required?: boolean;
    exampleValues?: string[];
}

export interface CommandFlag {
    name: string;
    description: string;
    short?: string;
    takesValue?: boolean;
    exampleValues?: string[];
}

// Advanced type inference for flags
export type FlagType<F extends CommandFlag> = F["takesValue"] extends true ? string : boolean;

export type FlagsType<Flags extends readonly CommandFlag[]> = {
    [K in Flags[number] as K["name"]]: FlagType<K>;
};

export interface CommandContext<Flags extends readonly CommandFlag[] = CommandFlag[]> {
    flags: FlagsType<Flags>;
    options: string[];
}

export interface CommandDefinition<Flags extends readonly CommandFlag[] = CommandFlag[]> {
    name: string;
    aliases?: string[];
    description: string;
    options?: CommandOption[];
    flags?: Flags;
    exampleUsage?: string[];
    execute: (ctx: CommandContext<Flags>) => Promise<void> | void;
}

export class Command<Flags extends readonly CommandFlag[] = CommandFlag[]> {
    name: string;
    aliases: string[];
    description: string;
    options: CommandOption[];
    flags: Flags;
    exampleUsage: string[];
    execute: (ctx: CommandContext<Flags>) => Promise<void> | void;

    constructor(def: CommandDefinition<Flags>) {
        this.name = def.name;
        this.aliases = def.aliases || [];
        this.description = def.description;
        this.options = def.options || [];
        this.flags = (def.flags || []) as Flags;
        this.exampleUsage = def.exampleUsage || [];
        this.execute = def.execute;
    }
}

// Command registry (global, for auto-registration)
export const commands: Command<any>[] = [];

export function registerCommand<Flags extends readonly CommandFlag[]>(def: CommandDefinition<Flags>): Command<Flags> {
    const cmd = new Command(def);
    commands.push(cmd);
    return cmd;
}
