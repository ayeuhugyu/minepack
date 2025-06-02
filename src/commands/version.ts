import { Command, registerCommand } from "../lib/command";
import chalk from "chalk";
import fs from "fs-extra";
import path from "path";

function getVersion() {
    // Try to read from package.json
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
