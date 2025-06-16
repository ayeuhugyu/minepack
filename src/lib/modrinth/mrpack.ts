import type { Pack } from "../pack";

export class MRPackFile {
    path: string;
    hashes: {
        sha1: string;
        sha512: string;
    }
    env: {
        client: "required" | "optional" | "unsupported";
        server: "required" | "optional" | "unsupported";
    }
    downloads: string[]; // a string of download URLs, we'll just include one though.
    fileSize: number; // in bytes
    dependencies: Record<"minecraft" | "forge" | "neoforge" | "fabric-loader" | "quilt-loader", string>; // version strings, ex. "minecraft": "1.20.1" or "fabric-loader": "0.16.14"

    constructor(data: MRPackFile) {
        this.path = data.path;
        this.hashes = data.hashes;
        this.env = data.env;
        this.downloads = data.downloads;
        this.fileSize = data.fileSize;
        this.dependencies = data.dependencies;
    }
}

export class MRPackIndex { // the modrinth.index.json file
    formatVersion: number = 1;
    game: string = "minecraft";
    name: string;
    summary?: string;
    files: MRPackFile[]

    constructor(data: { name: string; summary?: string; files: MRPackFile[] }) {
        this.name = data.name;
        this.files = data.files;
    }
}

export async function fromPack(pack: Pack) {
    const files: MRPackFile[] = [];
    const stubs = await pack.getStubs();
    for (const stub of stubs) {
        const dependencies: Record<string, string> = {
            minecraft: stub.gameVersion,
        }
        if ((pack.modloader.name === "fabric") || (pack.modloader.name === "quilt")) {
            dependencies[`${pack.modloader.name}-loader`] = pack.modloader.version;
        } else if ((pack.modloader.name === "forge") || (pack.modloader.name === "neoforge")) {
            dependencies[pack.modloader.name] = pack.modloader.version;
        } // liteloader users can just suffer 🤣
        files.push(new MRPackFile({
            path: stub.download.path,
            hashes: {
                sha1: stub.hashes.sha1,
                sha512: stub.hashes.sha512,
            },
            env: {
                client: stub.environments.client,
                server: stub.environments.server,
            },
            downloads: [stub.download.url],
            fileSize: stub.download.size,
            dependencies: dependencies
        }));
    }

    return new MRPackIndex({
        name: pack.name,
        summary: `"${pack.name}" created by ${pack.author} using Minepack:\n${pack.description}`,
        files: files
    });
}