import { registerCommand } from "../lib/command";
import { promptUser, selectFromList, statusMessage } from "../lib/util";
import fs from "fs";
import pathModule from "path";
import chalk from "chalk";
import { Pack } from "../lib/pack";
import { ModLoaders } from "../lib/loaderVersions";

registerCommand({
    name: "init",
    aliases: ["new"],
    description: "Initializes a new minepack project in the current directory, or a specified directory. ",
    options: [
        {
            name: "path",
            description: "The optional path to initialize the project in. Defaults to the current working directory.",
            required: false,
            exampleValues: ["C:\\Users\\User\\my-minepack-project", "~/my-minepack-project", "./my-minepack-project"],
        }
    ],
    flags: [
        {
            name: "force",
            description: "Force reinitialization of the project, even if it already exists.",
            short: "f",
            takesValue: false,
        },
        {
            name: "verbose",
            description: "Enable verbose output for debugging purposes.",
            short: "v",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack init",
        "minepack init ./my-minepack-project",
        "minepack init ~/my-minepack-project",
        "minepack init C:\\Users\\User\\my-minepack-project"
    ],
    execute: async ({ flags, options }) => {
        const inputPath = options[0] || process.cwd();
        const path = pathModule.isAbsolute(inputPath) ? inputPath : pathModule.resolve(process.cwd(), inputPath);
        if (flags.verbose) console.log(chalk.gray(`Initializing a new minepack project in ${chalk.yellowBright(path)}`));

        // Check if the directory already exists
        if (!fs.existsSync(path)) {
            fs.mkdirSync(path);
            if (flags.verbose) console.log(chalk.gray(`Created directory: ${chalk.yellowBright(path)}`));
        }
        if (Pack.isPack(path) && !flags.force) {
            console.error(chalk.redBright.bold(" ✖  This directory is already a minepack project. Pass --force to reinitialize it."));
            return;
        }
        if (Pack.isPack(path) && flags.force) {
            console.log(chalk.yellowBright.bold(" ⚠  Reinitializing existing minepack project..."));
        }

        const name = await promptUser(chalk.blueBright.bold("Pack name:") + chalk.reset());
        const description = await promptUser(chalk.blueBright.bold("Pack description:") + chalk.reset());
        const author = await promptUser(chalk.blueBright.bold("Author name:") + chalk.reset());
        let gameversion: string;
        while (true) {
            gameversion = await promptUser(chalk.blueBright.bold("Game version ") + chalk.gray("(e.g. 1.20.1):") + chalk.reset());
            if (gameversion.split('.').every(part => /^\d+$/.test(part)) && gameversion.split('.').length === 3) {
                break;
            }
            console.error(chalk.redBright.bold(" ✖  Invalid game version format. Please use a format like 1.20.1."));
        }

        const modloaderList = [
            "fabric",
            "quilt",
            "forge",
            "neoforge",
            "liteloader",
        ];
        const modloader = await selectFromList(
            modloaderList.map(key => chalk.greenBright(key)),
            chalk.blueBright.bold("Select a modloader:") + chalk.reset(),
        );

        const versionStatus = await statusMessage(chalk.gray("Fetching latest modloader version..."));
        const loaderData = ModLoaders[modloaderList[modloader]];
        if (!loaderData) {
            console.error(chalk.redBright.bold(" ✖  Invalid modloader selected."));
            versionStatus.done();
            return;
        }
        let modloaderVersion: string | undefined;
        try {
            modloaderVersion = (await loaderData.versionListGetter(gameversion))[1];
            if (!modloaderVersion) {
                throw new Error("No compatible version found.");
            }
        } catch (error: any) {
            versionStatus.update(
                chalk.redBright.bold(" ✖  Failed to fetch modloader version: ") +
                chalk.whiteBright(error?.message || error)
            );
            return;
        }
        versionStatus.done();

        const userInputtedModloaderVersion = (await promptUser(
            chalk.blueBright.bold(`Modloader version `) +
            chalk.gray(`(leave blank to use ${chalk.yellowBright(loaderData.friendlyName)} ${chalk.yellowBright(modloaderVersion)})`) +
            ":" + chalk.reset()
        )) || modloaderVersion;

        const pack: Pack = new Pack(
            name,
            author,
            description,
            path,
            {
                name: loaderData.name,
                version: userInputtedModloaderVersion
            },
            gameversion
        );

        if (flags.verbose) console.log(chalk.gray(`Creating pack file at ${chalk.yellowBright(path + "/pack.mp.json")}`));
        if (flags.force && Pack.isPack(path)) {
            console.log(chalk.yellowBright.bold(" ⚠  Overwriting existing pack file..."));
        }

        if (!fs.existsSync(path + "/tracked.mp.json")) {
            if (flags.verbose) console.log(chalk.gray(`Creating tracked files list at ${chalk.yellowBright(path + "/tracked.mp.json")}`));
            fs.writeFileSync(path + "/tracked.mp.json", JSON.stringify([]));
        }

        pack.write(flags.verbose);
        console.log(chalk.greenBright.bold(" ✔  Successfully initialized new minepack project!"));
    }
});