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

// Define interfaces for download and update
export interface ModDownload {
    url: string;
    hashFormat: HashFormat;
    hash: string;
}

export interface ModUpdate {
    modId: string;
    version: string;
}

export interface ModData {
    name: string;
    filename: string;
    side?: ModSide;
    download: {
        url: string;
        'hash-format': HashFormat | string;
        hash: string;
    };
    update?: {
        'mod-id': string;
        version: string;
    };
    dependencies?: string[];
}

export class Mod {
    name: string;
    filename: string;
    side: ModSide;
    download: ModDownload;
    update?: ModUpdate;
    dependencies: string[];

    constructor(data: ModData) {
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
    }
}

export function findMod(mods: (ModData & { _filename?: string })[], userInput: string): { mod: (ModData & { _filename?: string }) | null, matches: (ModData & { _filename?: string })[], fuzzy: boolean } {
    // 1. Direct filename match
    let mod = mods.find(m => m._filename === userInput || m._filename === userInput + ".json");
    if (mod) return { mod, matches: [], fuzzy: false };
    // 2. Lowercased filename match
    mod = mods.find(m => m._filename?.toLowerCase() === userInput.toLowerCase() || m._filename?.toLowerCase() === (userInput + ".json").toLowerCase());
    if (mod) return { mod, matches: [], fuzzy: false };
    // 3. By mod filename
    mod = mods.find(m => m.filename === userInput);
    if (mod) return { mod, matches: [], fuzzy: false };
    // 4. By download url
    mod = mods.find(m => m.download?.url === userInput);
    if (mod) return { mod, matches: [], fuzzy: false };
    // 5. Fuzzy search (by name, filename, url)
    const input = userInput.toLowerCase();
    const scored = mods.map(m => {
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