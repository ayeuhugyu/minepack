import { registerCommand } from "../lib/command";
import { findProject } from "../lib/modrinth/search";
import { projectToStub } from "../lib/modrinth/projectToStub";
import { Pack } from "../lib/pack";
import { Stub } from "../lib/stub";
import chalk from "chalk";
import { promptUser, selectFromList, multiSelectFromList } from "../lib/util";

registerCommand({
    name: "add",
    aliases: ["addmod"],
    description: "Add a mod (stub) to the current minepack project, with dependency resolution.",
    options: [
        {
            name: "mod",
            description: "The mod name, slug, or ID to add.",
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
        "minepack add sodium",
        "minepack add lithium --verbose"
    ],
    execute: async ({ flags, options }) => {
        const modQuery = options.join(" ").trim();
        if (!modQuery) {
            console.error(chalk.redBright.bold(" ✖  Please provide a mod to add."));
            return;
        }
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const verbose = !!flags.verbose;
        const existingStubs = pack.getStubs(verbose);
        // Search for the mod
        const project = await findProject(modQuery, pack, verbose);
        if (!project) {
            console.error(chalk.redBright.bold(" ✖  Could not find a matching mod on Modrinth."));
            return;
        }
        if (existingStubs.some(stub => stub.projectId === project.id || stub.slug === project.slug)) {
            console.log(chalk.yellowBright.bold(` ⚠  Mod '${project.slug}' is already added to this pack.`));
            return;
        }
        // Create stub and write it
        let stub;
        try {
            stub = await projectToStub(project, pack);
        } catch (err: any) {
            console.error(chalk.redBright.bold(` ✖  Failed to create stub: ${err.message}`));
            return;
        }
        stub.write(cwd, verbose);
        console.log(chalk.greenBright.bold(` ✔  Added '${project.title}' (${project.slug}) to the pack!`));
        // Handle dependencies
        if (stub.dependencies && stub.dependencies.length > 0) {
            // Filter out already present dependencies
            const missingDeps = stub.dependencies.filter(depId => !existingStubs.some(stub => stub.projectId === depId || stub.slug === depId));
            if (missingDeps.length > 0) {
                // Get project info for each dependency (if possible)
                const depProjects = [];
                for (const depId of missingDeps) {
                    const depProject = await findProject(depId, pack, verbose);
                    depProjects.push({ depId, depProject });
                }
                // List dependencies to user
                console.log(chalk.yellowBright.bold("The following dependencies were found:"));
                depProjects.forEach(({ depId, depProject }, idx) => {
                    if (depProject) {
                        console.log(`  ${chalk.gray(idx + 1 + ".")} ${chalk.bold(depProject.title)} (${chalk.yellowBright(depProject.slug)})`);
                    } else {
                        console.log(`  ${chalk.gray(idx + 1 + ".")} ${chalk.yellowBright(depId)} ${chalk.redBright("(not found on Modrinth)")}`);
                    }
                });
                // Ask user how to proceed
                async function askDepPrompt(): Promise<"all"|"none"|"select"> {
                    const depPrompt = await promptUser(chalk.yellowBright("Add all dependencies? (Y = all, N = none, S = select): "));
                    if (!depPrompt || depPrompt.trim().toLowerCase().startsWith("y")) return "all";
                    if (depPrompt.trim().toLowerCase().startsWith("n")) return "none";
                    if (depPrompt.trim().toLowerCase().startsWith("s")) return "select";
                    console.error(chalk.redBright.bold(" ✖  Please enter Y, N, or S."));
                    return askDepPrompt();
                }
                let toAdd: typeof depProjects = [];
                const depChoice = await askDepPrompt();
                if (depChoice === "all") {
                    toAdd = depProjects.filter(({ depProject }) => !!depProject);
                } else if (depChoice === "select") {
                    // Use multiSelectFromList for selection
                    const depTitles = depProjects
                        .filter(({ depProject }) => !!depProject)
                        .map(({ depProject }) => `${(depProject as any).title} (${(depProject as any).slug})`);
                    let selected: number[] = [];
                    if (depTitles.length > 0) {
                        selected = await multiSelectFromList(depTitles, chalk.blueBright("Select dependencies to add (toggle with numbers, 'a' for all, 'd' for done): "));
                    }
                    toAdd = depProjects.filter(({ depProject }, i) => !!depProject && selected.includes(i));
                } // else, none
                for (const { depProject } of toAdd) {
                    if (!depProject) continue;
                    try {
                        const depStub = await projectToStub(depProject, pack);
                        depStub.write(cwd, verbose);
                        console.log(chalk.greenBright.bold(` ✔  Added dependency '${depProject.title}' (${depProject.slug})!`));
                    } catch (err: any) {
                        console.error(chalk.redBright.bold(` ✖  Failed to add dependency '${depProject.slug}': ${err.message}`));
                    }
                }
            }
        }
    }
});
