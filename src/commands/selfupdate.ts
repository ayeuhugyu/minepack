import fs from "fs-extra";
import path from "path";
import os from "os";
import chalk from "chalk";
import https from "https";
import { Command, registerCommand } from "../lib/command";

const REPO = "ayeuhugyu/minepack";

function getPlatformBinaryName() {
    const platform = os.platform();
    const arch = os.arch();
    if (platform === "win32") return arch === "x64" ? "minepack-win-x64.exe" : "minepack.exe";
    if (platform === "darwin") return arch === "arm64" ? "minepack-mac-arm64" : "minepack-mac-x64";
    if (platform === "linux") {
        if (arch === "arm64") return "minepack-linux-arm64";
        if (arch === "x64") return "minepack-linux-x64";
    }
    return null;
}

async function fetchLatestRelease() {
    return new Promise((resolve, reject) => {
        https.get({
            hostname: "api.github.com",
            path: `/repos/${REPO}/releases/latest`,
            headers: { "User-Agent": "minepack-selfupdate" }
        }, res => {
            let data = "";
            res.on("data", chunk => data += chunk);
            res.on("end", () => {
                if (res.statusCode === 200) {
                    resolve(JSON.parse(data));
                } else {
                    reject(new Error(`Failed to fetch release: ${res.statusCode}`));
                }
            });
        }).on("error", reject);
    });
}

async function downloadFile(url: string, dest: string): Promise<void> {
    const res = await fetch(url, { redirect: "follow" });
    if (!res.ok || !res.body) throw new Error(`Failed to download: ${res.status} ${res.statusText}`);
    const fileStream = fs.createWriteStream(dest);
    const total = Number(res.headers.get("content-length")) || 0;
    let received = 0;
    const reader = res.body.getReader();
    let lastPercent = -1;
    await new Promise<void>(async (resolve, reject) => {
        function pump() {
            reader.read().then(({ done, value }) => {
                if (done) {
                    process.stdout.write("\r\n");
                    fileStream.end();
                    resolve();
                    return;
                }
                fileStream.write(Buffer.from(value), () => {
                    received += value.length;
                    if (total > 0) {
                        const percent = Math.floor((received / total) * 100);
                        if (percent !== lastPercent) {
                            process.stdout.write(`\r[download] ${percent}% (${received}/${total} bytes)`);
                            lastPercent = percent;
                        }
                    } else {
                        process.stdout.write(`\r[download] ${received} bytes`);
                    }
                    pump();
                });
            }).catch(reject);
        }
        fileStream.on("error", reject);
        pump();
    });
}

const selfupdateCommand = new Command({
    name: "selfupdate",
    description: "Attempt to update minepack to the latest release from GitHub.",
    arguments: [],
    flags: [
        { name: "version", aliases: ["v"], description: "Update to a specific version/tag", takesValue: true }
    ],
    async execute(_args, flags) {
        const binaryName = getPlatformBinaryName();
        if (!binaryName) {
            console.log(chalk.red("Unsupported platform/arch for self-update."));
            return;
        }
        let release;
        try {
            if (flags.version) {
                // Fetch specific release by tag
                release = await new Promise((resolve, reject) => {
                    https.get({
                        hostname: "api.github.com",
                        path: `/repos/${REPO}/releases/tags/${flags.version}`,
                        headers: { "User-Agent": "minepack-selfupdate" }
                    }, res => {
                        let data = "";
                        res.on("data", chunk => data += chunk);
                        res.on("end", () => {
                            if (res.statusCode === 200) {
                                resolve(JSON.parse(data));
                            } else {
                                reject(new Error(`Failed to fetch release: ${res.statusCode}`));
                            }
                        });
                    }).on("error", reject);
                });
            } else {
                release = await fetchLatestRelease();
            }
        } catch (err) {
            console.log(chalk.red(`Failed to fetch release info: ${err}`));
            return;
        }
        const rel = release as { assets?: any[] };
        const asset = (rel.assets || []).find((a: any) => a.name === binaryName);
        if (!asset) {
            console.log(chalk.red(`No binary found for your platform (${binaryName}) in the latest release.`));
            return;
        }
        const currentPath = process.execPath;
        const tmpPath = currentPath + ".tmp";
        console.log(chalk.gray(`[info] Downloading ${asset.name} from ${asset.browser_download_url}...`));
        try {
            await downloadFile(asset.browser_download_url, tmpPath);
        } catch (err) {
            console.log(chalk.red(`Failed to download binary: ${err}`));
            return;
        }
        try {
            // On Windows, can't overwrite running exe. On Unix, can usually overwrite.
            if (process.platform === "win32") {
                // Move current exe to .old, move tmp to exe, prompt user to restart
                const ext = path.extname(currentPath)
                const oldPath = ext
                    ? currentPath.slice(0, -ext.length) + "-old" + ext + ".old"
                    : currentPath + "-old";
                fs.renameSync(currentPath, oldPath);
                fs.renameSync(tmpPath, currentPath);
                console.log(chalk.green(`minepack updated! (Previous version saved as ${oldPath})`));
            } else {
                fs.renameSync(tmpPath, currentPath);
                fs.chmodSync(currentPath, 0o755);
                console.log(chalk.green("minepack updated!"));
            }
        } catch (err) {
            console.log(chalk.red(`Failed to replace binary: ${err}`));
            console.log(chalk.yellow(`The new binary is at: ${tmpPath}\nYou may need to replace the current executable manually.`));
            return;
        }
    }
});

registerCommand(selfupdateCommand);

export { selfupdateCommand };
