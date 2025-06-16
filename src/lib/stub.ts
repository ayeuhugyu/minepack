import fs from "fs";
import chalk from "chalk";

export class Stub {
    name: string;
    projectId: string;
    slug: string;
    type: "mod" | "resourcepack" | "shader";

    loader: string;
    gameVersion: string;
    
    hashes: {
        sha1: string;
        sha512: string;
    }
    download: {
        versionId: string;
        url: string;
        path: string;
        size: number;
    }
    environments: {
        client: "required" | "optional" | "unsupported";
        server: "required" | "optional" | "unsupported";
    }
    dependencies: string[]; // string of mod slugs

    constructor(data: {
        name: string;
        projectId: string;
        slug: string;
        type: "mod" | "resourcepack" | "shader";
        loader: string;
        gameVersion: string;
        hashes: { sha1: string; sha512: string };
        download: { versionId: string; url: string; path: string; size: number };
        environments: { client: "required" | "optional" | "unsupported"; server: "required" | "optional" | "unsupported" };
        dependencies: string[];
    }) {
        this.name = data.name;
        this.projectId = data.projectId;
        this.slug = data.slug;
        this.type = data.type;
        this.loader = data.loader;
        this.gameVersion = data.gameVersion;
        this.hashes = data.hashes;
        this.download = data.download;
        this.environments = data.environments;
        this.dependencies = data.dependencies;
    }

    write(rootPath: string, verbose: boolean = false): void {
        const stubFile = `${rootPath}/stubs/${this.name}.mp.json`;
        // ensure stubs directory
        if (!fs.existsSync(`${rootPath}/stubs`)) {
            if (verbose) console.log(chalk.gray(`Creating stubs directory at ${chalk.yellowBright(`${rootPath}/stubs`)}`));
            fs.mkdirSync(`${rootPath}/stubs`, { recursive: true });
        }
        fs.writeFileSync(stubFile, JSON.stringify(this, null, 2));
        if (verbose) {
            console.log(chalk.gray(`Wrote stub file to ${stubFile}`));
        }
    }

    static fromFile(rootPath: string, stubName: string, verbose: boolean = false): Stub | null {
        const fullPath = `${rootPath}/stubs/${stubName}.mp.json`;
        if (!fs.existsSync(fullPath)) {
            if (verbose) console.log(chalk.redBright.bold(` ✖  Stub file not found at ${fullPath}`));
            return null;
        }
        try {
            const rawData = fs.readFileSync(fullPath, "utf-8");
            const data = JSON.parse(rawData);
            return new Stub({
                name: data.name,
                projectId: data.projectId,
                slug: data.slug,
                type: data.type,
                loader: data.loader,
                gameVersion: data.gameVersion,
                hashes: {
                    sha1: data.hashes.sha1,
                    sha512: data.hashes.sha512
                },
                download: {
                    versionId: data.download.versionId,
                    url: data.download.url,
                    path: data.download.path,
                    size: data.download.size
                },
                environments: {
                    client: data.environments.client,
                    server: data.environments.server
                },
                dependencies: data.dependencies || []
            })
        } catch (error) {
            if (verbose) console.error(chalk.redBright.bold(` ✖  Failed to parse stub file at ${fullPath}:`, error));
            return null;
        }
    }
}