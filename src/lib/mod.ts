import chalk from 'chalk';

// Content types supported by minepack
export enum ContentType {
    Mod = 'mod',
    Resourcepack = 'resourcepack',
    Shaderpack = 'shaderpack',
    Datapack = 'datapack',
    Plugin = 'plugin',
    Unknown = 'unknown'
}

// Map Modrinth project_type to ContentType
export function modrinthTypeToContentType(projectType: string): ContentType {
    switch (projectType) {
        case 'mod': return ContentType.Mod;
        case 'resourcepack': return ContentType.Resourcepack;
        case 'shader': return ContentType.Shaderpack;
        case 'datapack': return ContentType.Datapack;
        case 'plugin': return ContentType.Plugin;
        default: return ContentType.Unknown;
    }
}

// Map ContentType to folder name
export function contentTypeToFolder(type: ContentType): string {
    switch (type) {
        case ContentType.Mod: return 'mods';
        case ContentType.Resourcepack: return 'resourcepacks';
        case ContentType.Shaderpack: return 'shaderpacks';
        case ContentType.Datapack: return 'datapacks';
        case ContentType.Plugin: return 'plugins';
        default: return 'other';
    }
}

// Define enums for side and hash format
export enum ModSide {
    Client = 'client',
    Server = 'server',
    Both = 'both'
}

export enum HashFormat {
    Sha1 = 'sha1',
    Sha256 = 'sha256',
    Sha512 = 'sha512',
    Md5 = 'md5'
    // Add more as needed
}

// Generic content data model
export interface ContentData {
    type: ContentType;
    name: string;
    filename: string;
    side?: ModSide;
    download: {
        url: string;
        'hash-format': HashFormat | string;
        hash: string;
    };
    update?: {
        'mod-id'?: string;
        version?: string;
    };
    dependencies?: string[];
    fileSize: number; // now required, always set (0 if unknown)
    // Allow extra fields for future extensibility
    [key: string]: any;
}

// Backward compatibility: treat ModData as ContentData with type 'mod'
export type ModData = ContentData;

// Update Mod class to Content class
export class Content {
    type: ContentType;
    name: string;
    filename: string;
    side: ModSide;
    download: {
        url: string;
        hashFormat: HashFormat | string;
        hash: string;
    };
    update?: {
        modId?: string;
        version?: string;
    };
    dependencies: string[];
    [key: string]: any;

    constructor(data: ContentData) {
        this.type = data.type;
        this.name = data.name;
        this.filename = data.filename;
        this.side = data.side || ModSide.Both;
        this.download = {
            url: data.download.url,
            hashFormat: data.download['hash-format'] as HashFormat,
            hash: data.download.hash
        };
        if (data.update) {
            this.update = {
                modId: data.update['mod-id'],
                version: data.update.version
            };
        }
        this.dependencies = data.dependencies ?? [];
        // Copy any extra fields
        Object.assign(this, data);
    }
}

// Update findMod to findContent, but keep findMod for backward compatibility
export function findContent(contents: (ContentData & { _filename?: string })[], userInput: string): { mod: (ContentData & { _filename?: string }) | null, matches: (ContentData & { _filename?: string })[], fuzzy: boolean } {
    // 1. Direct filename match
    let mod = contents.find(m => m._filename === userInput || m._filename === userInput + ".json");
    if (mod) return { mod, matches: [], fuzzy: false };
    // 2. Lowercased filename match
    mod = contents.find(m => m._filename?.toLowerCase() === userInput.toLowerCase() || m._filename?.toLowerCase() === (userInput + ".json").toLowerCase());
    if (mod) return { mod, matches: [], fuzzy: false };
    // 3. By filename
    mod = contents.find(m => m.filename === userInput);
    if (mod) return { mod, matches: [], fuzzy: false };
    // 4. By download url
    mod = contents.find(m => m.download?.url === userInput);
    if (mod) return { mod, matches: [], fuzzy: false };
    // 5. Fuzzy search (by name, filename, url)
    const input = userInput.toLowerCase();
    const scored = contents.map(m => {
        let score = 0;
        if (m.name?.toLowerCase().includes(input)) score += 3;
        if (m.filename?.toLowerCase().includes(input)) score += 2;
        if (m.download?.url?.toLowerCase().includes(input)) score += 1;
        return { mod: m, score };
    }).filter(x => x.score > 0).sort((a, b) => b.score - a.score);
    if (scored.length) {
        return { mod: null, matches: scored.slice(0, 5).map(x => x.mod), fuzzy: true };
    }
    return { mod: null, matches: [], fuzzy: false };
}

// For backward compatibility
export function findMod(mods: (ModData & { _filename?: string })[], userInput: string): { mod: (ModData & { _filename?: string }) | null, matches: (ModData & { _filename?: string })[], fuzzy: boolean } {
    console.log(chalk.gray(`[info] Searching for: ${userInput}`));
    // 1. Direct filename match (json or jar)
    let mod = mods.find(m => m._filename === userInput || m._filename === userInput + ".json" || m._filename === userInput + ".jar");
    if (mod) {
        console.log(chalk.gray(`[info] Direct filename match: ${mod._filename}`));
        return { mod, matches: [], fuzzy: false };
    }
    // 2. Lowercased filename match
    mod = mods.find(m => m._filename?.toLowerCase() === userInput.toLowerCase() || m._filename?.toLowerCase() === (userInput + ".json").toLowerCase() || m._filename?.toLowerCase() === (userInput + ".jar").toLowerCase());
    if (mod) {
        console.log(chalk.gray(`[info] Lowercased filename match: ${mod._filename}`));
        return { mod, matches: [], fuzzy: false };
    }
    // 3. By mod filename (json or jar)
    mod = mods.find(m => m.filename === userInput || m.filename === userInput + ".jar");
    if (mod) {
        console.log(chalk.gray(`[info] Mod filename match: ${mod.filename}`));
        return { mod, matches: [], fuzzy: false };
    }
    // 4. By download url
    mod = mods.find(m => m.download?.url === userInput);
    if (mod) {
        console.log(chalk.gray(`[info] Download URL match: ${mod.download?.url}`));
        return { mod, matches: [], fuzzy: false };
    }
    // 5. Fuzzy search (by name, filename, url, _filename)
    const input = userInput.toLowerCase();
    const scored = mods.map(m => {
        let score = 0;
        if (m.name?.toLowerCase().includes(input)) score += 3;
        if (m.filename?.toLowerCase().includes(input)) score += 2;
        if (m._filename?.toLowerCase().includes(input)) score += 2;
        if (m.download?.url?.toLowerCase().includes(input)) score += 1;
        return { mod: m, score };
    }).filter(x => x.score > 0).sort((a, b) => b.score - a.score);
    if (scored.length) {
        console.log(chalk.gray(`[info] Fuzzy matches found: ${scored.slice(0, 5).map(x => x.mod.name || x.mod._filename).join(", ")}`));
        return { mod: null, matches: scored.slice(0, 5).map(x => x.mod), fuzzy: true };
    }
    console.log(chalk.gray(`[info] No matches found.`));
    return { mod: null, matches: [], fuzzy: false };
}