import { registerCommand } from "../lib/command";
import { VERSION } from "../version";
import chalk from "chalk";

registerCommand({
    name: "version",
    aliases: [],
    description: "Returns the current version of minepack.",
    options: [],
    flags: [],
    exampleUsage: ["minepack version"],
    execute: async ({ flags, options }) => {
        console.log(`minepack ${chalk.blue(VERSION)}`);
    }
});