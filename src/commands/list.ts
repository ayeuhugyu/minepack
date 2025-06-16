import { registerCommand } from "../lib/command";
import { Pack } from "../lib/pack";
import chalk from "chalk";
import prettyBytes from "pretty-bytes";

registerCommand({
    name: "list",
    aliases: ["ls", "mods"],
    description: "List all mods in the current minepack project.",
    options: [
        {
            name: "detail",
            description: "Level of detail: 'basic', 'full', or 'json'.",
            required: false,
            exampleValues: ["basic", "full", "json"],
        }
    ],
    flags: [
        {
            name: "hashes",
            description: "Show hashes (sha1/sha512) for each mod.",
            short: "h",
            takesValue: false,
        },
        {
            name: "urls",
            description: "Show download URLs for each mod.",
            short: "u",
            takesValue: false,
        },
        {
            name: "env",
            description: "Show environment support (client/server).",
            short: "e",
            takesValue: false,
        },
        {
            name: "ids",
            description: "Show project IDs.",
            short: "i",
            takesValue: false,
        },
        {
            name: "size",
            description: "Show file size for each mod.",
            short: "s",
            takesValue: false,
        }
    ],
    exampleUsage: [
        "minepack list",
        "minepack list full --hashes --urls",
        "minepack ls --env --ids"
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
        const detail = (options[0] || "basic").toLowerCase();
        stubs.forEach(stub => {
            let line = chalk.bold.cyanBright(stub.name) +
                chalk.gray(" (") + chalk.yellowBright(stub.slug) + chalk.gray(") ") +
                chalk.magentaBright(`[${stub.type.toUpperCase()}]`);
            console.log(line);
            if (detail === "full" || flags.hashes || flags.urls || flags.env || flags.ids || flags.size) {
                // Calculate max key length for alignment
                const keyLabels = [
                    (flags.size || detail === "full") ? "Size" : null,
                    (flags.hashes || detail === "full") ? "SHA1" : null,
                    (flags.hashes || detail === "full") ? "SHA512" : null,
                    (flags.urls || detail === "full") ? "URL" : null,
                    (flags.env || detail === "full") ? "Environments" : null,
                    (flags.ids || detail === "full") ? "Project ID" : null
                ].filter(Boolean) as string[];
                const maxKeyLen = Math.max(...keyLabels.map(k => k.length));
                function padKey(label: string) {
                    return label.padEnd(maxKeyLen, ' ');
                }
                if (flags.size || detail === "full") {
                    console.log("  " + chalk.blueBright(` ${padKey("Size")} `) + " " + chalk.blueBright(prettyBytes(stub.download.size)));
                }
                if (flags.hashes || detail === "full") {
                    console.log("  " + chalk.greenBright(` ${padKey("SHA1")} `) + " " + chalk.greenBright(stub.hashes.sha1));
                    console.log("  " + chalk.green(` ${padKey("SHA512")} `) + " " + chalk.green(stub.hashes.sha512));
                }
                if (flags.urls || detail === "full") {
                    console.log("  " + chalk.yellowBright(` ${padKey("URL")} `) + " " + chalk.yellowBright(stub.download.url));
                }
                if (flags.env || detail === "full") {
                    console.log("  " + chalk.magentaBright(` ${padKey("Environments")} `) +
                        ` ` + chalk.magentaBright("client:") + " " + chalk.whiteBright(stub.environments.client) +
                        ", " + chalk.magentaBright("server:") + " " + chalk.whiteBright(stub.environments.server));
                }
                if (flags.ids || detail === "full") {
                    console.log("  " + chalk.cyanBright(` ${padKey("Project ID")} `) + " " + chalk.cyanBright(stub.projectId));
                }
            }
            if (detail === "json") {
                console.log(chalk.gray(JSON.stringify(stub, null, 2)));
            }
        });
    }
});
