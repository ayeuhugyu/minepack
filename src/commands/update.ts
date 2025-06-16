import { registerCommand } from "../lib/command";
import { findProject } from "../lib/modrinth/search";
import { projectToStub } from "../lib/modrinth/projectToStub";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import { promptUser, selectFromList, multiSelectFromList } from "../lib/util";

registerCommand({
    name: "update",
    aliases: ["updatemod", "upgrade"],
    description: "Update a mod (stub) in the current minepack project, or all mods with --all.",
    options: [
        {
            name: "mod",
            description: "The mod name, slug, or ID to update. Omit if using --all.",
            required: false,
            exampleValues: ["sodium", "lithium", "modrinth-xyz123"],
        }
    ],
    flags: [
        {
            name: "all",
            description: "Update all mods in the pack.",
            short: "a",
            takesValue: false,
        },
        {
            name: "verbose",
            description: "Enable verbose output.",
            short: "v",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack update sodium",
        "minepack update --all",
        "minepack update lithium --verbose"
    ],
    execute: async ({ flags, options }) => {
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const verbose = !!flags.verbose;
        const stubs = pack.getStubs(verbose);
        if (flags.all) {
            if (stubs.length === 0) {
                console.log(chalk.yellowBright.bold("No mods to update."));
                return;
            }
            console.log(chalk.cyanBright.bold(`Updating all mods in the pack...`));
            await Promise.all(stubs.map(async (stub) => {
                const project = await findProject(stub.slug, pack, verbose);
                if (!project) {
                    console.log(chalk.redBright.bold(` ✖  Could not find mod '${stub.slug}' on Modrinth. Skipping.`));
                    return;
                }
                try {
                    const newStub = await projectToStub(project, pack);
                    newStub.write(cwd, verbose);
                    console.log(chalk.greenBright.bold(` ✔  Updated '${project.title}' (${project.slug})!`));
                } catch (err: any) {
                    console.error(chalk.redBright.bold(` ✖  Failed to update '${project.slug}': ${err.message}`));
                }
            }));
            return;
        }
        // Single mod update
        const modQuery = options.join(" ").trim();
        if (!modQuery) {
            console.error(chalk.redBright.bold(" ✖  Please provide a mod to update, or use --all."));
            return;
        }
        // Try to find the stub
        let target = stubs.find(stub => stub.slug.toLowerCase() === modQuery.toLowerCase());
        if (!target) target = stubs.find(stub => stub.projectId.toLowerCase() === modQuery.toLowerCase());
        if (!target) target = stubs.find(stub => stub.name.toLowerCase() === modQuery.toLowerCase());
        if (!target) {
            // Fuzzy match
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
        // Update the stub
        const project = await findProject(target.slug, pack, verbose);
        if (!project) {
            console.error(chalk.redBright.bold(` ✖  Could not find mod '${target.slug}' on Modrinth.`));
            return;
        }
        try {
            const newStub = await projectToStub(project, pack);
            newStub.write(cwd, verbose);
            console.log(chalk.greenBright.bold(` ✔  Updated '${project.title}' (${project.slug})!`));
        } catch (err: any) {
            console.error(chalk.redBright.bold(` ✖  Failed to update '${project.slug}': ${err.message}`));
        }
    }
});
