import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import { Stub } from "../lib/stub";
import chalk from "chalk";
import { promptUser, selectFromList, multiSelectFromList } from "../lib/util";
import fs from "fs";

registerCommand({
    name: "remove",
    aliases: ["removemod", "rm"],
    description: "Remove a mod (stub) from the current minepack project, with dependency checks and fuzzy matching.",
    options: [
        {
            name: "mod",
            description: "The mod slug, projectId, or name to remove.",
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
        "minepack remove sodium",
        "minepack rm lithium --verbose"
    ],
    execute: async ({ flags, options }) => {
        const modQuery = options.join(" ").trim();
        if (!modQuery) {
            console.error(chalk.redBright.bold(" ✖  Please provide a mod to remove."));
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
            console.error(chalk.redBright.bold(" ✖  No mods found in this pack."));
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
                    console.log(chalk.redBright.bold(" ✖  No matching mod found. Please manually delete the stub file in the stubs directory."));
                    return;
                }
            } else {
                console.log(chalk.redBright.bold(" ✖  No matching mod found. Please manually delete the stub file in the stubs directory."));
                return;
            }
        }
        // Check for dependents
        const dependents = stubs.filter(stub => stub.dependencies && (stub.dependencies.includes(target.slug) || stub.dependencies.includes(target.projectId)));
        if (dependents.length > 0) {
            console.log(chalk.yellowBright.bold(" ⚠  The following mods depend on this mod:"));
            dependents.forEach(dep => {
                console.log(`  ${chalk.bold(dep.name)} (${chalk.yellowBright(dep.slug)})`);
            });
            const depNames = dependents.map(dep => `${chalk.bold(dep.name)} (${chalk.yellowBright(dep.slug)})`);
            const options = [
                ...depNames,
                chalk.redBright("Cancel (do not remove anything)")
            ];
            const selected = await multiSelectFromList(options, chalk.yellowBright("Select any dependent mods you would also like to remove, or choose cancel:"));
            if (selected.includes(options.length - 1) || selected.length === 0) {
                console.log(chalk.gray("Aborted mod removal."));
                return;
            }
            // Remove selected dependents
            for (const idx of selected) {
                if (idx < dependents.length) {
                    const dep = dependents[idx];
                    const depStubFile = `${cwd}/stubs/${dep.name}.mp.json`;
                    if (fs.existsSync(depStubFile)) {
                        fs.unlinkSync(depStubFile);
                        console.log(chalk.greenBright.bold(` ✔  Removed dependent '${dep.name}' (${dep.slug}) from the pack!`));
                    }
                }
            }
        }
        // Remove stub file
        const stubFile = `${cwd}/stubs/${target.name}.mp.json`;
        if (fs.existsSync(stubFile)) {
            fs.unlinkSync(stubFile);
            console.log(chalk.greenBright.bold(` ✔  Removed '${target.name}' (${target.slug}) from the pack!`));
        } else {
            console.log(chalk.redBright.bold(` ✖  Stub file not found at ${stubFile}.`));
        }

        // Orphaned dependency check
        if (target.dependencies && target.dependencies.length > 0) {
            // Find stubs for each dependency
            const orphanCandidates = target.dependencies
                .map(depId => stubs.find(stub => stub.projectId === depId || stub.slug === depId))
                .filter(Boolean) as Stub[];
            // For each, check if any other stub depends on it
            const trulyOrphaned = orphanCandidates.filter(depStub =>
                !stubs.some(s => s !== target && s.dependencies && (s.dependencies.includes(depStub.slug) || s.dependencies.includes(depStub.projectId)))
            );
            if (trulyOrphaned.length > 0) {
                console.log(chalk.yellowBright.bold("The following mods were dependencies of the removed mod, and are no longer required by any other mod:"));
                trulyOrphaned.forEach((stub, i) => {
                    console.log(`  ${chalk.gray(i + 1 + ".")} ${chalk.bold(stub.name)} (${chalk.yellowBright(stub.slug)})`);
                });
                const orphanTitles = trulyOrphaned.map(stub => `${stub.name} (${stub.slug})`);
                const selected = await multiSelectFromList(orphanTitles, chalk.blueBright("Select any orphaned dependencies you would also like to remove (toggle with numbers, 'a' for all, 'd' for done): "));
                for (const idx of selected) {
                    const stub = trulyOrphaned[idx];
                    const stubFile = `${cwd}/stubs/${stub.name}.mp.json`;
                    if (fs.existsSync(stubFile)) {
                        fs.unlinkSync(stubFile);
                        console.log(chalk.greenBright.bold(` ✔  Removed orphaned dependency '${stub.name}' (${stub.slug}) from the pack!`));
                    }
                }
            }
        }
    }
});
