import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import prettyBytes from "pretty-bytes";
import stripAnsi from "strip-ansi";

registerCommand({
    name: "map",
    aliases: ["depmap", "graph", "showmap"],
    description: "Display a colorized, aligned dependency map of the pack's mods.",
    options: [], // No options currently supported
    flags: [],
    exampleUsage: [
        "minepack map"
    ],
    execute: async ({ flags, options }) => {
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const stubs = pack.getStubs();
        if (stubs.length === 0) {
            console.log(chalk.yellowBright.bold("No mods found in this pack."));
            return;
        }
        // Build a map of dependency -> [dependents]
        const dependencyToDependents: Record<string, string[]> = {};
        stubs.forEach(stub => {
            (stub.dependencies || []).forEach(dep => {
                if (!dependencyToDependents[dep]) dependencyToDependents[dep] = [];
                dependencyToDependents[dep].push(stub.slug);
            });
        });
        // Find all unique dependencies (including those with no dependents)
        const allDependencySlugs = Array.from(new Set([
            ...Object.keys(dependencyToDependents),
            ...stubs.map(s => s.slug)
        ]));
        // Find mods with no dependencies
        const modsWithNoDeps = stubs.filter(stub => !stub.dependencies || stub.dependencies.length === 0);
        // Prepare column widths for alignment
        // Account for offset in dependent lines ("   ├─" or "   ╰─")
        const treeOffset = '   ├─'.length;
        const colData = stubs.map(stub => [
            stub.name,
            stub.slug,
            stub.type.toUpperCase(),
            stub.projectId,
            `C:${stub.environments.client[0].toUpperCase()} S:${stub.environments.server[0].toUpperCase()}`,
            prettyBytes(stub.download.size)
        ]);
        const colWidths = [0, 0, 0, 0, 0, 0];
        colData.forEach(cols => cols.forEach((val, i) => { if (val.length > colWidths[i]) colWidths[i] = val.length; }));
        // For dependent lines, add treeOffset to the first column
        const depColWidths = [colWidths[0] + treeOffset, ...colWidths.slice(1)];
        // Helper to format a mod line
        function formatModLine(stub: typeof stubs[0], isDependent = false) {
            const widths = isDependent ? depColWidths : colWidths;
            // Add extra spaces after the name for dependent lines to offset the tree characters
            const namePad = isDependent ? ' '.repeat(treeOffset) : '';
            return (
                chalk.blueBright.bold(` ${stub.name}${namePad}`.padEnd(widths[0])) +
                chalk.gray(' (') + chalk.yellowBright(stub.slug.padEnd(widths[1])) + chalk.gray(') ')
                + chalk.magentaBright(`[${stub.type.toUpperCase().padEnd(widths[2])}]`)
                + chalk.gray(' | ') + chalk.bold('ID:') + ' ' + chalk.cyanBright(stub.projectId.padEnd(widths[3]))
                + chalk.gray(' | ') + chalk.bold('Env:') + ' ' + chalk.greenBright(`C:${stub.environments.client[0].toUpperCase()} S:${stub.environments.server[0].toUpperCase()}`.padEnd(widths[4]))
                + chalk.gray(' | ') + chalk.bold('Size:') + ' ' + chalk.whiteBright(prettyBytes(stub.download.size).padEnd(widths[5]))
            );
        }
        // Helper to format a dependency node
        function formatDepNode(depStub: typeof stubs[0] | undefined, depId: string) {
            return depStub
                ? chalk.yellowBright(` ${depStub.name.padEnd(colWidths[0])} `)
                    + chalk.gray(' (') + chalk.yellowBright(depStub.slug.padEnd(colWidths[1])) + chalk.gray(') ')
                    + chalk.magentaBright(`[${depStub.type.toUpperCase().padEnd(colWidths[2])}]`)
                    + chalk.gray(' | ') + chalk.bold('ID:') + ' ' + chalk.cyanBright(depStub.projectId.padEnd(colWidths[3]))
                    + chalk.gray(' | ') + chalk.bold('Env:') + ' ' + chalk.greenBright(`C:${depStub.environments.client[0].toUpperCase()} S:${depStub.environments.server[0].toUpperCase()}`.padEnd(colWidths[4]))
                    + chalk.gray(' | ') + chalk.bold('Size:') + ' ' + chalk.whiteBright(prettyBytes(depStub.download.size).padEnd(colWidths[5]))
                : chalk.redBright(' [Missing stub] ') + chalk.yellowBright(depId);
        }
        // Calculate the max line length for the box lines
        let maxLineLen = 0;
        // Helper to build a line with dynamic column alignment, given a prefix (tree chars or empty)
        function buildAlignedLine({
            prefix = '',
            name,
            slug,
            type,
            projectId,
            env,
            size,
            nameColor = (x: string) => x,
            slugColor = (x: string) => x,
            typeColor = (x: string) => x,
            idColor = (x: string) => x,
            envColor = (x: string) => x,
            sizeColor = (x: string) => x,
            extraNamePad = 0,
        }: {
            prefix?: string,
            name: string,
            slug: string,
            type: string,
            projectId: string,
            env: string,
            size: string,
            nameColor?: (x: string) => string,
            slugColor?: (x: string) => string,
            typeColor?: (x: string) => string,
            idColor?: (x: string) => string,
            envColor?: (x: string) => string,
            sizeColor?: (x: string) => string,
            extraNamePad?: number,
        }) {
            // Build the prefix+name segment and measure its visible length
            const prefixName = prefix + nameColor(name);
            // For dependency nodes, add extraNamePad (treeOffset) to the name padding
            const namePad = Math.max(0, colWidths[0] - stripAnsi(name).length + (extraNamePad || 0));
            let line = prefixName + ' '.repeat(namePad) + '  ';
            line += slugColor(slug) + ' '.repeat(colWidths[1] - stripAnsi(slug).length) + '  ';
            line += typeColor(type) + ' '.repeat(colWidths[2] - stripAnsi(type).length) + '  ';
            line += idColor(projectId) + ' '.repeat(colWidths[3] - stripAnsi(projectId).length) + '  ';
            line += envColor(env) + ' '.repeat(colWidths[4] - stripAnsi(env).length) + '  ';
            line += sizeColor(size) + ' '.repeat(colWidths[5] - stripAnsi(size).length);
            return line;
        }
        // Build all lines for measurement and output
        let allLines: string[] = [];
        // 1. Dependencies with dependents
        allDependencySlugs.forEach(depId => {
            const depStub = stubs.find(s => s.slug === depId || s.projectId === depId);
            const dependents = dependencyToDependents[depId] || [];
            if (dependents.length > 0) {
                // Dependency node: add treeOffset to the name padding
                if (depStub) {
                    allLines.push(buildAlignedLine({
                        prefix: '',
                        name: depStub.name,
                        slug: depStub.slug,
                        type: depStub.type.toUpperCase(),
                        projectId: depStub.projectId,
                        env: `C:${depStub.environments.client[0].toUpperCase()} S:${depStub.environments.server[0].toUpperCase()}`,
                        size: prettyBytes(depStub.download.size),
                        nameColor: chalk.yellowBright,
                        slugColor: chalk.yellowBright,
                        typeColor: chalk.magentaBright,
                        idColor: chalk.cyanBright,
                        envColor: chalk.greenBright,
                        sizeColor: chalk.whiteBright,
                        extraNamePad: treeOffset,
                    }));
                } else {
                    allLines.push(chalk.redBright(' [Missing stub] ') + chalk.yellowBright(depId));
                }
                // Dependents
                dependents.forEach((slug, idx) => {
                    const branch = (idx === dependents.length - 1) ? chalk.gray('   ╰─') : chalk.gray('   ├─');
                    const stubObj = stubs.find(s => s.slug === slug);
                    if (stubObj) {
                        allLines.push(buildAlignedLine({
                            prefix: branch,
                            name: stubObj.name,
                            slug: stubObj.slug,
                            type: stubObj.type.toUpperCase(),
                            projectId: stubObj.projectId,
                            env: `C:${stubObj.environments.client[0].toUpperCase()} S:${stubObj.environments.server[0].toUpperCase()}`,
                            size: prettyBytes(stubObj.download.size),
                            nameColor: chalk.blueBright.bold,
                            slugColor: chalk.yellowBright,
                            typeColor: chalk.magentaBright,
                            idColor: chalk.cyanBright,
                            envColor: chalk.greenBright,
                            sizeColor: chalk.whiteBright,
                        }));
                    } else {
                        allLines.push(branch + chalk.redBright(' [Missing stub] ') + chalk.yellowBright(slug));
                    }
                });
                allLines.push('');
            }
        });
        // 2. Mods with no dependencies
        if (modsWithNoDeps.length > 0) {
            allLines.push(chalk.gray.whiteBright(' No dependencies '));
            modsWithNoDeps.forEach((stub, idx) => {
                const branch = (idx === modsWithNoDeps.length - 1) ? chalk.gray('   ╰─') : chalk.gray('   ├─');
                allLines.push(buildAlignedLine({
                    prefix: branch,
                    name: stub.name,
                    slug: stub.slug,
                    type: stub.type.toUpperCase(),
                    projectId: stub.projectId,
                    env: `C:${stub.environments.client[0].toUpperCase()} S:${stub.environments.server[0].toUpperCase()}`,
                    size: prettyBytes(stub.download.size),
                    nameColor: chalk.blueBright.bold,
                    slugColor: chalk.yellowBright,
                    typeColor: chalk.magentaBright,
                    idColor: chalk.cyanBright,
                    envColor: chalk.greenBright,
                    sizeColor: chalk.whiteBright,
                }));
            });
        }
        // Find the max line length
        maxLineLen = Math.max(...allLines.map(l => stripAnsi(l).length));
        // Calculate total size
        const totalSize = stubs.reduce((sum, stub) => sum + (stub.download?.size || 0), 0);
        // Print pack data header
        const packDataLine =
            chalk.magentaBright.bold(' Pack:') +
            ' ' + chalk.bold.whiteBright(pack.name) +
            chalk.gray(' | ') + chalk.bold('Author:') + ' ' + chalk.cyanBright(pack.author) +
            chalk.gray(' | ') + chalk.bold('Game:') + ' ' + chalk.yellowBright(pack.gameVersion) +
            chalk.gray(' | ') + chalk.bold('Modloader:') + ' ' + chalk.greenBright(`${pack.modloader.name} ${pack.modloader.version}`) +
            chalk.gray(' | ') + chalk.bold('Total Size:') + ' ' + chalk.whiteBright(prettyBytes(totalSize)) + " ";
        const packLineLen = stripAnsi(packDataLine).length;
        const lineLen = Math.max(maxLineLen, packLineLen);
        console.log(chalk.gray('─'.repeat(lineLen)));
        console.log(packDataLine.padEnd(lineLen + (packDataLine.length - stripAnsi(packDataLine).length)));
        console.log(chalk.gray('─'.repeat(lineLen)));
        // Print all lines
        allLines.forEach(l => console.log(l.padEnd(lineLen + (l.length - stripAnsi(l).length))));
        console.log(chalk.gray('─'.repeat(lineLen)));
    }
});
