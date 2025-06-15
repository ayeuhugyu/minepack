import fs from "fs";
import chalk from "chalk";

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

    constructor(name: string, author: string, description: string, rootPath: string, modloader: { name: string; version: string }, gameVersion: string) {
        this.name = name;
        this.author = author;
        this.description = description;
        this.rootPath = rootPath;
        this.modloader = modloader;
        this.gameVersion = gameVersion;
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

            return new Pack(
                packData.name,
                packData.author,
                packData.description,
                path,
                packData.modloader,
                packData.gameVersion
            );
        } catch (error) {
            console.error(chalk.redBright.bold(` ✖  Failed to parse pack at ${path}:`, error));
            return null;
        }
    }

    write(verbose: boolean = false): void {
        const packFile = `${this.rootPath}/pack.mp.json`;
        fs.writeFileSync(packFile, JSON.stringify(this, null, 2));
        if (verbose) {
            console.log(chalk.gray(`Writing pack file to ${packFile}...`));
        }
        console.log(chalk.greenBright.bold(` ✔  Pack file written to ${packFile}`));
    }

    getTrackedFiles(verbose: boolean = false): string[] {
        const trackedFilePath = `${this.rootPath}/tracked.mp.json`;
        if (verbose) {
            console.log(chalk.gray(`Reading tracked files from ${trackedFilePath}...`));
        }
        let tracked;
        try {
            tracked = JSON.parse(fs.readFileSync(trackedFilePath, 'utf-8'));
        } catch (err) {
            console.error(chalk.redBright.bold(` ✖  Failed to read tracked files: ${err}`));
            return [];
        }
        if (!Array.isArray(tracked)) {
            console.error(chalk.redBright.bold(" ✖  Invalid tracked files format. Expected an array."));
            return [];
        }
        const files = tracked.map((file: string) => {
            if (typeof file !== "string") {
                if (verbose) {
                    console.error(chalk.redBright.bold(" ✖  Invalid tracked file format. Expected a string."));
                }
                return "";
            }
            if (verbose) {
                console.log(chalk.gray(`Tracked file: ${file}`));
            }
            return file;
        }).filter((file: string) => file !== "");
        if (verbose) {
            console.log(chalk.gray(`Total tracked files: ${files.length}`));
        }
        return files;
    }

    addTrackedFile(file: string, verbose: boolean = false): void {
        const trackedFiles = this.getTrackedFiles(verbose);
        if (trackedFiles.includes(file)) {
            console.warn(chalk.yellowBright.bold(` ⚠  File ${file} is already tracked.`));
            return;
        }
        trackedFiles.push(file);
        fs.writeFileSync(`${this.rootPath}/tracked.mp.json`, JSON.stringify(trackedFiles, null, 2));
        if (verbose) {
            console.log(chalk.gray(`Tracked files updated. New count: ${trackedFiles.length}`));
            console.log(chalk.greenBright.bold(` ✔  Added ${file} to tracked files.`));
        }
    }
}