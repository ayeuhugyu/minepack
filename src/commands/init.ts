import { registerCommand } from "../lib/command";
import { promptUser, selectFromList, statusMessage } from "../lib/util";
import fs from "fs";
import path from "path";
import chalk from "chalk";
import { Pack } from "../lib/pack";
import { ModLoaders } from "../lib/loaderVersions";
import { projectReadmeContent } from "../lib/projectReadmeContent";

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
        const projectPath = path.isAbsolute(inputPath) ? inputPath : path.resolve(process.cwd(), inputPath);
        if (flags.verbose) console.log(chalk.gray(`Initializing a new minepack project in ${chalk.yellowBright(projectPath)}`));

        // Check if the directory already exists
        if (!fs.existsSync(projectPath)) {
            fs.mkdirSync(projectPath);
            if (flags.verbose) console.log(chalk.gray(`Created directory: ${chalk.yellowBright(projectPath)}`));
        }
        if (Pack.isPack(projectPath) && !flags.force) {
            console.error(chalk.redBright.bold(" ✖  This directory is already a minepack project. Pass --force to reinitialize it."));
            return;
        }
        if (Pack.isPack(projectPath) && flags.force) {
            console.log(chalk.yellowBright.bold(" ⚠  Reinitializing existing minepack project..."));
        }

        const name = await promptUser(chalk.blueBright.bold("Pack name:") + chalk.reset());
        const description = await promptUser(chalk.blueBright.bold("Pack description:") + chalk.reset());
        const author = await promptUser(chalk.blueBright.bold("Author name:") + chalk.reset());
        const version = (await promptUser(chalk.blueBright.bold("Pack version ") + chalk.gray("(default 1.0.0):") + chalk.reset())) || "1.0.0";
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
            projectPath,
            {
                name: loaderData.name,
                version: userInputtedModloaderVersion
            },
            gameversion,
            version
        );

        if (flags.verbose) console.log(chalk.gray(`Creating pack file at ${chalk.yellowBright(projectPath + "/pack.mp.json")}`));
        if (flags.force && Pack.isPack(projectPath)) {
            console.log(chalk.yellowBright.bold(" ⚠  Overwriting existing pack file..."));
        }

        const stubsDir = path.join(projectPath, "stubs");
        if (!fs.existsSync(stubsDir)) {
            if (flags.verbose) console.log(chalk.gray(`Creating stubs directory at ${chalk.yellowBright(stubsDir)}`));
            fs.mkdirSync(stubsDir);
        }

        const overridesDir = path.join(projectPath, "overrides");
        if (!fs.existsSync(overridesDir)) {
            if (flags.verbose) console.log(chalk.gray(`Creating overrides directory at ${chalk.yellowBright(overridesDir)}`));
            fs.mkdirSync(overridesDir);
        }

        // Create a README.md file if it doesn't exist
        const readmePath = path.join(projectPath, "README.md");
        if (!fs.existsSync(readmePath)) {
            if (flags.verbose) console.log(chalk.gray(`Creating README.md at ${chalk.yellowBright(readmePath)}`));
            fs.writeFileSync(
                readmePath,
                projectReadmeContent
            );
        }

        pack.write(flags.verbose);
        console.log(chalk.greenBright.bold(" ✔  Successfully initialized new minepack project!\nSee the README.md file created within the project for a little more information on how to use minepack."));
    }
});