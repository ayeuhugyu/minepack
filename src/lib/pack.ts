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
        const stubsDir = `${this.rootPath}/stubs`;
        let files: string[] = [];
        try {
            if (!fs.existsSync(stubsDir)) {
                if (verbose) {
                    console.log(chalk.gray(`Stubs directory does not exist at ${stubsDir}. It will be created.`));
                }
                fs.mkdirSync(stubsDir, { recursive: true });
                return [];
            }
            files = fs.readdirSync(stubsDir)
                .filter(file => file.endsWith('.mp.json'))
                .map(file => `${stubsDir}/${file}`);
            if (verbose) {
                files.forEach(file => console.log(chalk.gray(`Tracked file: ${file}`)));
                console.log(chalk.gray(`Total tracked files: ${files.length}`));
            }
        } catch (err) {
            console.error(chalk.redBright.bold(` ✖  Failed to read stubs directory: ${err}`));
            return [];
        }
        return files;
    }

    addTrackedFile(file: string, verbose: boolean = false): void {
        const stubsDir = `${this.rootPath}/stubs`;
        if (!fs.existsSync(stubsDir)) {
            fs.mkdirSync(stubsDir, { recursive: true });
            if (verbose) {
                console.log(chalk.gray(`Created stubs directory at ${stubsDir}.`));
            }
        }
        const fileName = file.endsWith('.mp.json') ? file : `${file}.mp.json`;
        const filePath = `${stubsDir}/${fileName}`;
        if (fs.existsSync(filePath)) {
            console.warn(chalk.yellowBright.bold(` ⚠  File ${filePath} is already tracked.`));
            return;
        }
        fs.writeFileSync(filePath, '{}');
        if (verbose) {
            console.log(chalk.gray(`Tracked files updated. New count: ${this.getTrackedFiles().length + 1}`));
            console.log(chalk.greenBright.bold(` ✔  Added ${filePath} to tracked files.`));
        }
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