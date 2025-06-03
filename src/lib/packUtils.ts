import path from "path";
import fs from "fs-extra";

export const CONTENT_FOLDERS = ["mods", "resourcepacks", "shaderpacks", "datapacks", "plugins"];
export const STUB_EXT = ".mp.json";
export const TRACKED_FILE = "tracked.mp.json";

export function getContentFolders(): string[] {
    return CONTENT_FOLDERS;
}

export function getStubFilesFromTracked(rootDir: string): string[] {
    const trackedPath = path.join(rootDir, TRACKED_FILE);
    if (!fs.existsSync(trackedPath)) return [];
    return JSON.parse(fs.readFileSync(trackedPath, "utf-8"));
}

export function updateTrackedFile(rootDir: string, stubFiles: string[]): void {
    const trackedPath = path.join(rootDir, TRACKED_FILE);
    fs.writeFileSync(trackedPath, JSON.stringify(stubFiles, null, 2));
}

export function findAllStubFiles(rootDir: string): string[] {
    const stubs: string[] = [];
    for (const folder of CONTENT_FOLDERS) {
        const dir = path.join(rootDir, folder);
        if (!fs.existsSync(dir)) continue;
        for (const file of fs.readdirSync(dir)) {
            if (file.endsWith(STUB_EXT)) {
                stubs.push(path.join(folder, file));
            }
        }
    }
    return stubs;
}

export function addStubToTracked(rootDir: string, stubPath: string): void {
    const tracked = getStubFilesFromTracked(rootDir);
    if (!tracked.includes(stubPath)) {
        tracked.push(stubPath);
        updateTrackedFile(rootDir, tracked);
    }
}

export function removeStubFromTracked(rootDir: string, stubPath: string): void {
    let tracked = getStubFilesFromTracked(rootDir);
    tracked = tracked.filter(f => f !== stubPath);
    updateTrackedFile(rootDir, tracked);
}
