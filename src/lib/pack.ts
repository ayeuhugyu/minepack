import fs from "fs";
import chalk from "chalk";
import { Stub } from "./stub"; // Adjust the import path as necessary

export class Pack {
    name: string;
    author: string;
    description: string;
    rootPath: string;
    modloader: {
        name: string;
        version: string;
    }
    gameVersion: string;
    version: string;

    constructor(name: string, author: string, description: string, rootPath: string, modloader: { name: string; version: string }, gameVersion: string, version: string = "1.0.0") {
        this.name = name;
        this.author = author;
        this.description = description;
        this.rootPath = rootPath;
        this.modloader = modloader;
        this.gameVersion = gameVersion;
        this.version = version;
    }

    static isPack(path: string): boolean {
        const packFile = `${path}/pack.mp.json`;
        if (!fs.existsSync(packFile)) {
            return false;
        } else {
            return true;
        }
    }

    static parse(path: string): Pack | null {
        try {
            const isPack = Pack.isPack(path);
            if (!isPack) {
                throw new Error(`No pack file could be found`);
            }
            const packFile = `${path}/pack.mp.json`;
            const rawData = fs.readFileSync(packFile, 'utf-8');
            const packData = JSON.parse(rawData);

            // Basic type checks
            if (typeof packData.name !== "string") {
                throw new Error("Invalid or missing property: name");
            }
            if (typeof packData.author !== "string") {
                throw new Error("Invalid or missing property: author");
            }
            if (typeof packData.description !== "string") {
                throw new Error("Invalid or missing property: description");
            }
            if (typeof packData.modloader !== "object" || packData.modloader === null) {
                throw new Error("Invalid or missing property: modloader");
            }
            if (typeof packData.modloader.name !== "string") {
                throw new Error("Invalid or missing property: modloader.name");
            }
            if (typeof packData.modloader.version !== "string") {
                throw new Error("Invalid or missing property: modloader.version");
            }
            if (typeof packData.gameVersion !== "string") {
                throw new Error("Invalid or missing property: gameVersion");
            }
            // Version is optional, default to 1.0.0
            const version = typeof packData.version === "string" ? packData.version : "1.0.0";

            return new Pack(
                packData.name,
                packData.author,
                packData.description,
                path,
                packData.modloader,
                packData.gameVersion,
                version
            );
        } catch (error) {
            console.error(chalk.redBright.bold(` ✖  Failed to parse pack at ${path}:`, error));
            return null;
        }
    }

    write(verbose: boolean = false): void {
        const packFile = `${this.rootPath}/pack.mp.json`;
        // Write version property
        fs.writeFileSync(packFile, JSON.stringify(this, null, 2));
        if (verbose) {
            console.log(chalk.gray(`Writing pack file to ${packFile}...`));
        }
        console.log(chalk.greenBright.bold(` ✔  Pack file written to ${packFile}`));
    }

    getStubs(verbose: boolean = false): Stub[] {
        const stubsDir = `${this.rootPath}/stubs`;
        let stubs: Stub[] = [];
        try {
            if (!fs.existsSync(stubsDir)) {
                if (verbose) {
                    console.log(chalk.gray(`Stubs directory does not exist at ${stubsDir}.`));
                }
                return [];
            }
            const files = fs.readdirSync(stubsDir)
                .filter(file => file.endsWith('.mp.json'));
            for (const file of files) {
                const stub = Stub.fromFile(this.rootPath, file.replace(/\.mp\.json$/, ""), verbose);
                if (stub) stubs.push(stub);
            }
            if (verbose) {
                stubs.forEach(stub => console.log(chalk.gray(`Stub: ${stub.name} (${stub.projectId})`)));
                console.log(chalk.gray(`Total stubs: ${stubs.length}`));
            }
        } catch (err) {
            console.error(chalk.redBright.bold(` ✖  Failed to read stubs directory: ${err}`));
            return [];
        }
        return stubs;
    }
}