import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import { selectFromList } from "../lib/util";
import prettyBytes from "pretty-bytes";

registerCommand({
    name: "query",
    aliases: ["hasmod", "inpack"],
    description: "Query if a mod is present in the current minepack project.",
    options: [
        {
            name: "mod",
            description: "The mod slug, projectId, or name to query.",
            required: true,
            exampleValues: ["sodium", "lithium", "modrinth-xyz123"],
        }
    ],
    flags: [
        {
            name: "verbose",
            description: "Enable verbose output.",
            short: "v",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack query sodium",
        "minepack query lithium --verbose"
    ],
    execute: async ({ flags, options }) => {
        const modQuery = options.join(" ").trim();
        if (!modQuery) {
            console.error(chalk.redBright.bold(" ✖  Please provide a mod to query."));
            return;
        }
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const verbose = !!flags.verbose;
        const stubs = pack.getStubs(verbose);
        if (stubs.length === 0) {
            console.log(chalk.yellowBright.bold("No mods found in this pack."));
            return;
        }
        // 1. Try slug match
        let target = stubs.find(stub => stub.slug.toLowerCase() === modQuery.toLowerCase());
        // 2. Try projectId match
        if (!target) target = stubs.find(stub => stub.projectId.toLowerCase() === modQuery.toLowerCase());
        // 3. Try exact lowercase name match
        if (!target) target = stubs.find(stub => stub.name.toLowerCase() === modQuery.toLowerCase());
        // 4. Fuzzy substring match (top 5)
        if (!target) {
            const matches = stubs
                .map(stub => ({
                    stub,
                    score: (stub.name.toLowerCase().includes(modQuery.toLowerCase()) ? 1 : 0)
                        + (stub.slug.toLowerCase().includes(modQuery.toLowerCase()) ? 1 : 0)
                        + (stub.projectId.toLowerCase().includes(modQuery.toLowerCase()) ? 1 : 0)
                }))
                .filter(x => x.score > 0)
                .sort((a, b) => b.score - a.score)
                .slice(0, 5);
            if (matches.length > 0) {
                const list = matches.map(x => `${chalk.bold(x.stub.name)} (${chalk.yellowBright(x.stub.slug)})`);
                const idx = await selectFromList([...list, chalk.redBright("None of these")], chalk.yellowBright("No exact match found. Which mod did you mean?"));
                if (idx < matches.length) {
                    target = matches[idx].stub;
                } else {
                    console.log(chalk.redBright.bold(" ✖  No matching mod found in this pack."));
                    return;
                }
            } else {
                console.log(chalk.redBright.bold(" ✖  No matching mod found in this pack."));
                return;
            }
        }
        // Print info in a colorized, aligned style
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
        console.log(chalk.greenBright.bold(target.name) + chalk.gray(` (${target.slug})`));
        console.log(chalk.gray("Type: ") + chalk.magentaBright(target.type));
        console.log(chalk.gray("Project ID: ") + chalk.cyanBright(target.projectId));
        console.log(chalk.gray("Loader: ") + chalk.yellowBright(target.loader));
        console.log(chalk.gray("Game Version: ") + chalk.yellowBright(target.gameVersion));
        console.log(chalk.gray("Environments: ") + chalk.greenBright(`Client: ${target.environments.client}, Server: ${target.environments.server}`));
        console.log(chalk.gray("File Size: ") + chalk.blueBright(prettyBytes(target.download.size)));
        console.log(chalk.gray("SHA1: ") + chalk.greenBright(target.hashes.sha1));
        console.log(chalk.gray("SHA512: ") + chalk.green(target.hashes.sha512));
        console.log(chalk.gray("Download URL: ") + chalk.yellowBright(target.download.url));
        if (target.dependencies && target.dependencies.length > 0) {
            console.log(chalk.gray("Dependencies: ") + chalk.whiteBright(target.dependencies.join(", ")));
        } else {
            console.log(chalk.gray("Dependencies: ") + chalk.whiteBright("None"));
        }
        console.log(chalk.gray("────────────────────────────────────────────────────────────"));
    }
});
