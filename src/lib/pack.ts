export enum ModLoader {
    Fabric = 'fabric-loader',
    Forge = 'forge',
    Quilt = 'quilt-loader',
    NeoForge = 'neoforge',
}

export class PackMeta {
    name: string;
    version: string;
    author: string;
    gameversion: string;
    modloader: {
        name: ModLoader
        version: string;
    };

    constructor(data: {
        name: string;
        version: string;
        author: string;
        gameversion: string;
        modloader: {
            name: ModLoader;
            version: string;
        };
    }) {
        this.name = data.name;
        this.version = data.version;
        this.author = data.author;
        this.gameversion = data.gameversion;
        this.modloader = {
            name: data.modloader.name,
            version: data.modloader.version
        };
    }
}