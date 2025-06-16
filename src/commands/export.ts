import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import fs from "fs";
import path from "path";
import { fromPack, MRPackIndex } from "../lib/modrinth/mrpack";

function copyDirSync(src: string, dest: string) {
    if (!fs.existsSync(src)) return;
    if (!fs.existsSync(dest)) fs.mkdirSync(dest, { recursive: true });
    for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
        const srcPath = path.join(src, entry.name);
        const destPath = path.join(dest, entry.name);
        if (entry.isDirectory()) {
            copyDirSync(srcPath, destPath);
        } else {
            fs.copyFileSync(srcPath, destPath);
        }
    }
}

registerCommand({
    name: "export",
    aliases: ["mrpack", "to-mrpack"],
    description: "Export the pack as a Modrinth .mrpack file (modrinth.index.json + overrides).",
    options: [],
    flags: [
        {
            name: "required",
            description: "Force all environments to 'required' in the export. This can be useful for testing mods that are usually only needed on the server.",
            short: "r",
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
        "minepack export",
        "minepack export --required"
    ],
    execute: async ({ flags }) => {
        const cwd = process.cwd();
        const pack = Pack.parse(cwd);
        if (!pack) {
            console.error(chalk.redBright.bold(" ✖  Not a minepack project directory."));
            return;
        }
        const verbose = !!flags.verbose;
        const required = !!flags.required;
        const exportDir = path.join(pack.rootPath, ".exports");
        const overridesSrc = path.join(pack.rootPath, "overrides");
        const overridesDest = path.join(exportDir, "overrides");
        // Clean export dir if exists
        if (fs.existsSync(exportDir)) {
            fs.rmSync(exportDir, { recursive: true, force: true });
        }
        fs.mkdirSync(exportDir, { recursive: true });
        // Copy overrides
        if (fs.existsSync(overridesSrc)) {
            if (verbose) console.log(chalk.gray(`Copying overrides from ${overridesSrc} to ${overridesDest}`));
            copyDirSync(overridesSrc, overridesDest);
        } else if (verbose) {
            console.log(chalk.yellowBright("No overrides directory found, skipping."));
        }
        // Create modrinth.index.json
        if (verbose) console.log(chalk.gray("Building modrinth.index.json..."));
        const mrpackIndex: MRPackIndex = await fromPack(pack, { required });
        const indexPath = path.join(exportDir, "modrinth.index.json");
        fs.writeFileSync(indexPath, JSON.stringify(mrpackIndex, null, 2));
        if (verbose) console.log(chalk.greenBright.bold(` ✔  Wrote ${indexPath}`));
        // Zip the .exports directory
        const outName = `${pack.name.replace(/\s+/g, "_")}-${pack.modloader.name}.mrpack`;
        const outPath = path.join(pack.rootPath, outName);
        if (fs.existsSync(outPath)) fs.rmSync(outPath);
        // Use PowerShell's Compress-Archive if on Windows, otherwise use zip
        const isWin = process.platform === "win32";
        // Detect if running inside PowerShell
        const isPwsh = isWin && !!process.env.PSModulePath && !process.env.SHELL && !process.env.TERM_PROGRAM;
        let zipCmd;
        if (isWin) {
            if (isPwsh) {
                // Directly call Compress-Archive in PowerShell
                zipCmd = `Compress-Archive -Path "${exportDir}/*" -DestinationPath "${outPath}" -Force`;
            } else {
                // Use full path to powershell.exe to invoke Compress-Archive
                const powershellPath = process.env.SystemRoot
                    ? `${process.env.SystemRoot}\\System32\\WindowsPowerShell\\v1.0\\powershell.exe`
                    : "powershell";
                zipCmd = `\"${powershellPath}\" Compress-Archive -Path \"${exportDir}/*\" -DestinationPath \"${outPath}\" -Force`;
            }
        } else {
            zipCmd = `zip -r \"${outPath}\" .exports`;
        }
        const { execSync } = require("child_process");
        try {
            if (verbose) console.log(chalk.gray(`Zipping export directory to ${outPath}...`));
            execSync(zipCmd, { cwd: pack.rootPath, stdio: verbose ? "inherit" : "ignore" });
            console.log(chalk.greenBright.bold(` ✔  Exported pack to ${outPath}`));
        } catch (err: any) {
            console.error(chalk.redBright.bold(" ✖  Failed to zip export directory: \n", err));
            return;
        } finally {
            // Always clean up the .exports directory
            if (fs.existsSync(exportDir)) {
                try {
                    fs.rmSync(exportDir, { recursive: true, force: true });
                    if (verbose) console.log(chalk.gray(`Cleaned up temporary export directory: ${exportDir}`));
                } catch (cleanupErr) {
                    if (verbose) console.log(chalk.redBright(`Failed to clean up .exports: ${cleanupErr}`));
                }
            }
        }
    }
});
