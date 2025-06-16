import { registerCommand } from "../lib/command";
import { findProject } from "../lib/modrinth/search";
import chalk from "chalk";
import { Pack } from "../lib/pack";

registerCommand({
    name: "search",
    aliases: ["findmod"],
    description: "Search for a mod on Modrinth and display basic info.",
    options: [
        {
            name: "query",
            description: "The mod name, slug, or ID to search for.",
            required: true,
            exampleValues: ["sodium", "lithium", "modrinth-xyz123"],
        }
    ],
    flags: [
        {
            name: "gameVersion",
            description: "Override the game version to search for.",
            short: "g",
            takesValue: true,
        },
        {
            name: "modloader",
            description: "Override the modloader to search for.",
            short: "m",
            takesValue: true,
        },
        {
            name: "verbose",
            description: "Enable verbose output.",
            short: "v",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack search sodium",
        "minepack search lithium -g 1.20.1 -m fabric",
        "minepack search modrinth-xyz123 --verbose"
    ],
    execute: async ({ flags, options }) => {
        const query = options.join(" ").trim();
        if (!query) {
            console.error(chalk.redBright.bold(" ✖  Please provide a search query."));
            return;
        }

        // Try to use Pack from current directory, fallback to no pack context if not found
        let packData: Pack | undefined = undefined;
        try {
            const parsed = await Pack.parse(process.cwd());
            if (parsed) packData = parsed;
        } catch {
            // Ignore error, fallback to no pack context
        }
        
        const project = await findProject(query, packData, !!flags.verbose);
        if (!project) {
            console.log(chalk.redBright.bold(" ✖  No matching mod found."));
            return;
        }

        // Print basic info
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
        console.log(chalk.greenBright.bold(project.title) + chalk.gray(` (${project.slug})`));
        console.log(chalk.whiteBright(project.description));
        console.log(chalk.gray("Downloads: ") + chalk.yellowBright(project.downloads.toLocaleString()));
        console.log(chalk.gray("Project Type: ") + chalk.cyanBright(project.project_type));
        console.log(chalk.gray("Game Versions: ") + chalk.yellowBright(project.game_versions.join(", ")));
        console.log(chalk.gray("Loaders: ") + chalk.yellowBright(project.loaders.join(", ")));
        if (project.source_url) console.log(chalk.gray("Source: ") + chalk.blueBright(project.source_url));
        if (project.issues_url) console.log(chalk.gray("Issues: ") + chalk.blueBright(project.issues_url));
        if (project.wiki_url) console.log(chalk.gray("Wiki: ") + chalk.blueBright(project.wiki_url));
        if (project.discord_url) console.log(chalk.gray("Discord: ") + chalk.blueBright(project.discord_url));
        if (project.donation_url) console.log(chalk.gray("Donate: ") + chalk.blueBright(project.donation_url));
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
    }
});
