import { registerCommand } from "../lib/command";
import chalk from "chalk";
import fs from "fs";
import path from "path";
import https from "https";

const REPO = "ayeuhugyu/minepack";

function getPlatformBinaryName() {
    const platform = process.platform;
    const arch = process.arch;
    if (platform === "win32") return arch === "x64" ? "minepack-win-x64.exe" : "minepack.exe";
    if (platform === "darwin") return arch === "arm64" ? "minepack-mac-arm64" : "minepack-mac-x64";
    if (platform === "linux") {
        if (arch === "arm64") return "minepack-linux-arm64";
        if (arch === "x64") return "minepack-linux-x64";
    }
    return null;
}

async function fetchRelease(tag?: string): Promise<any> {
    return new Promise((resolve, reject) => {
        https.get({
            hostname: "api.github.com",
            path: tag ? `/repos/${REPO}/releases/tags/${tag}` : `/repos/${REPO}/releases/latest`,
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
                            process.stdout.write("\r" + chalk.gray("[download] ") + chalk.greenBright(`${percent}%`) + chalk.gray(` (${received}/${total} bytes)`));
                            lastPercent = percent;
                        }
                    } else {
                        process.stdout.write("\r" + chalk.gray("[download] ") + chalk.yellowBright(`${received} bytes`));
                    }
                    pump();
                });
            }).catch(reject);
        }
        fileStream.on("error", reject);
        pump();
    });
}

registerCommand({
    name: "selfupdate",
    aliases: ["upgrade", "update-self"],
    description: "Update minepack to the latest release from GitHub.",
    options: [],
    flags: [
        { name: "version", description: "Update to a specific version/tag", short: "v", takesValue: true }
    ],
    exampleUsage: [
        "minepack selfupdate",
        "minepack selfupdate --version v5"
    ],
    execute: async ({ flags }) => {
        const binaryName = getPlatformBinaryName();
        if (!binaryName) {
            console.log(chalk.redBright.bold("✖ Unsupported platform/arch for self-update."));
            return;
        }
        let release;
        try {
            release = await fetchRelease(flags.version);
        } catch (err: any) {
            console.log(chalk.redBright.bold(`✖ Failed to fetch release info: ${err.message || err}`));
            return;
        }
        const rel = release as { assets?: any[] };
        const asset = (rel.assets || []).find((a: any) => a.name === binaryName);
        if (!asset) {
            console.log(chalk.redBright.bold(`✖ No binary found for your platform (${binaryName}) in the release.`));
            return;
        }
        const currentPath = process.execPath;
        const tmpPath = currentPath + ".tmp";
        console.log(chalk.gray(`[info] Downloading ${chalk.cyanBright(asset.name)} from ${chalk.underline(asset.browser_download_url)}...`));
        try {
            await downloadFile(asset.browser_download_url, tmpPath);
        } catch (err: any) {
            console.log(chalk.redBright.bold(`✖ Failed to download binary: ${err.message || err}`));
            return;
        }
        try {
            if (process.platform === "win32") {
                const ext = path.extname(currentPath);
                const oldPath = ext
                    ? currentPath.slice(0, -ext.length) + "-old" + ext + ".old"
                    : currentPath + "-old";
                fs.renameSync(currentPath, oldPath);
                fs.renameSync(tmpPath, currentPath);
                console.log(chalk.greenBright.bold(`✔ minepack updated! (Previous version saved as ${oldPath})`));
            } else {
                fs.renameSync(tmpPath, currentPath);
                fs.chmodSync(currentPath, 0o755);
                console.log(chalk.greenBright.bold("✔ minepack updated!"));
            }
        } catch (err: any) {
            console.log(chalk.redBright.bold(`✖ Failed to replace binary: ${err.message || err}`));
            console.log(chalk.yellowBright(`The new binary is at: ${tmpPath}\nYou may need to replace the current executable manually.`));
            return;
        }
    }
});
