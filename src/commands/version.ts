import { Command, registerCommand } from "../lib/command";
import chalk from "chalk";
import fs from "fs-extra";
import path from "path";

// Use build-time injected version if available
const BUILT_VERSION = typeof process !== 'undefined' && process.env.MINEPACK_VERSION ? process.env.MINEPACK_VERSION : undefined;

function getVersion() {
    if (BUILT_VERSION) return BUILT_VERSION;
    // Try to read from package.json (dev mode only)
    try {
        const pkgPath = path.resolve(__dirname, "../../package.json");
        if (fs.existsSync(pkgPath)) {
            const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf-8"));
            return pkg.version || "unknown";
        }
    } catch {}
    return "unknown";
}

const versionCommand = new Command({
    name: "version",
    description: "Show the current minepack version.",
    arguments: [],
    flags: [],
    async execute() {
        const version = getVersion();
        console.log(chalk.green(`v${version}`));
    }
});

registerCommand(versionCommand);

export { versionCommand };
