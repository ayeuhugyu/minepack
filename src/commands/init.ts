import chalk from "chalk";
import path from "path";
import fs from "fs-extra";
import readline from "readline/promises";
import {
    Command,
    registerCommand,
    commands,
    globalFlags,
    getCommand
} from '../lib/command';
import { PackMeta } from "../lib/pack";

interface ProjectFile {
    path: string;
    content: string;
}

// Help command implementation
const initCommand = new Command({
    name: 'init',
    description: 'Initializes a new Minepack project in the current or specified directory.',
    flags: [
        {
            name: 'directory',
            aliases: ['d'],
            description: 'The directory to initialize the project in. Defaults to the current directory.',
            takesValue: true
        }
    ],
    examples: [
        {
            description: 'Initialize a new Minepack project',
            usage: 'minepack init'
        },
        {
            description: 'Initialize a new Minepack project in a specific directory',
            usage: 'minepack init --directory ./my-project'
        }
    ],
    async execute(args, flags) {
        console.log(chalk.bold("Initializing Minepack project..."));
        const dirInput = args.directory || flags.directory || '.';
        const dir = path.resolve(process.cwd(), dirInput.toString());
        console.log(`Project will be initialized in: ${chalk.cyan(dir)}`);

        if (!dirInput || dirInput === '.') {
            console.log(chalk.yellow("No directory specified, using current directory."));
        } else {
            console.log(`Using specified directory: ${chalk.cyan(dir)}`);
        }

        if (fs.existsSync(dir)) {
            console.log(chalk.yellow(`Directory ${chalk.cyan(dir)} already exists.`));
        } else {
            console.log(chalk.yellow(`Directory ${chalk.cyan(dir)} does not already exist, it will be created.`));
            fs.mkdirSync(dir, { recursive: true });
        }

        // === Detect existing pack.json ===
        const packJsonPath = path.join(dir, 'pack.json');
        if (fs.existsSync(packJsonPath)) {
            try {
                const raw = fs.readFileSync(packJsonPath, 'utf8');
                const parsed = JSON.parse(raw);
                // Try to construct PackMeta, will throw if invalid
                new PackMeta(parsed);
                console.log(chalk.red(`A Minepack project already exists in ${chalk.cyan(dir)}.`));
                console.log(chalk.red('Aborting initialization.'));
                return;
            } catch (e) {
                // If error, allow to continue (corrupt or not a Minepack pack.json)
                console.log(chalk.yellow('Existing pack.json is invalid or not a Minepack project. Continuing...'));
            }
        }

        // === Interactive prompts for PackMeta ===
        const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
        const name = await rl.question('Project name: ') || 'My Minepack Project';
        const version = await rl.question('Project version (e.g. 1.0.0): ') || '1.0.0';
        const author = await rl.question('Author: ') || 'Unknown Author';
        const gameversion = await rl.question('Minecraft version (e.g. 1.20.4): ') || '1.20.4';

        // Modloader selection
        const { ModLoaders } = await import('../lib/version');
        const modloaderNames = Object.keys(ModLoaders);
        console.log('Available modloaders:');
        modloaderNames.forEach((ml, i) => {
            console.log(`  [${i + 1}] ${ModLoaders[ml].friendlyName} (${ml})`);
        });
        let modloaderIndex = parseInt(await rl.question('Select modloader [number]: '), 10) - 1;
        while (isNaN(modloaderIndex) || modloaderIndex < 0 || modloaderIndex >= modloaderNames.length) {
            modloaderIndex = parseInt(await rl.question('Invalid selection. Select modloader [number]: '), 10) - 1;
        }
        const modloaderName = modloaderNames[modloaderIndex];
        const modloader = ModLoaders[modloaderName];
        // Fetch modloader versions
        let modloaderVersion = '';
        try {
            const [versions, latest] = await modloader.versionListGetter(gameversion);
            console.log(`Latest version for ${modloader.friendlyName}: ${latest}`);
            modloaderVersion = await rl.question(`Modloader version [default: ${latest}]: `) || latest;
        } catch (e) {
            console.log(chalk.red(`Could not fetch versions for ${modloader.friendlyName}: ${e}`));
            modloaderVersion = await rl.question('Modloader version: ');
        }

        await rl.close();

        // === Write pack.json ===
        const packMeta = new PackMeta({
            name,
            version,
            author,
            gameversion,
            modloader: {
                name: modloaderName as import("../lib/pack").ModLoader,
                version: modloaderVersion
            }
        });

        fs.writeFileSync(packJsonPath, JSON.stringify(packMeta, null, 4));
        console.log(chalk.green(`Created ${packJsonPath}`));

        // Create mods directory if not exists
        const modsDir = path.join(dir, 'mods');
        if (!fs.existsSync(modsDir)) {
            fs.mkdirSync(modsDir, { recursive: true });
            console.log(`Created directory: ${chalk.cyan(modsDir)}`);
        }

        console.log(chalk.green("Minepack project initialized successfully!"));
    }
});

registerCommand(initCommand);

// Export for main CLI
export { initCommand };