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

    constructor(data: MRPackFile) {
        this.path = data.path;
        this.hashes = data.hashes;
        this.env = data.env;
        this.downloads = data.downloads;
        this.fileSize = data.fileSize;
    }
}

export class MRPackIndex { // the modrinth.index.json file
    formatVersion: number = 1;
    game: string = "minecraft";
    name: string;
    summary?: string;
    files: MRPackFile[];
    versionId: string;
    dependencies: Partial<Record<"minecraft" | "forge" | "neoforge" | "fabric-loader" | "quilt-loader", string>>;

    constructor(data: { name: string; summary?: string; files: MRPackFile[]; versionId: string; dependencies: Partial<Record<"minecraft" | "forge" | "neoforge" | "fabric-loader" | "quilt-loader", string>> }) {
        this.name = data.name;
        this.summary = data.summary;
        this.files = data.files;
        this.versionId = data.versionId;
        this.dependencies = data.dependencies;
    }
}

export async function fromPack(pack: Pack, opts?: { required?: boolean }) {
    const files: MRPackFile[] = [];
    const stubs = await pack.getStubs();

    // Build dependencies once, for the whole pack
    const dependencies: Partial<Record<"minecraft" | "forge" | "neoforge" | "fabric-loader" | "quilt-loader", string>> = {
        minecraft: pack.gameVersion,
    };
    if ((pack.modloader.name === "fabric") || (pack.modloader.name === "quilt")) {
        dependencies[`${pack.modloader.name}-loader` as "fabric-loader" | "quilt-loader"] = pack.modloader.version;
    } else if ((pack.modloader.name === "forge") || (pack.modloader.name === "neoforge")) {
        dependencies[pack.modloader.name as "forge" | "neoforge"] = pack.modloader.version;
    }

    for (const stub of stubs) {
        files.push(new MRPackFile({
            path: stub.download.path,
            hashes: {
            sha1: stub.hashes.sha1,
            sha512: stub.hashes.sha512,
            },
            env: {
            client: opts?.required && stub.environments.client !== "unsupported" ? "required" : stub.environments.client,
            server: opts?.required && stub.environments.server !== "unsupported" ? "required" : stub.environments.server,
            },
            downloads: [stub.download.url],
            fileSize: stub.download.size,
        }));
    }

    return new MRPackIndex({
        name: pack.name,
        summary: `"${pack.name}" created by ${pack.author} using Minepack:\n${pack.description}`,
        files: files,
        versionId: pack.version || "1.0.0",
        dependencies: dependencies
    });
}