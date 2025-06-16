import type { Pack } from "../pack";
import type { ModrinthProject } from "./search";
import { BaseURL } from "./apiUrl";
import { Stub } from "../stub";

// Types for Modrinth API responses
export interface ModrinthVersionFile {
    hashes: { sha1: string; sha512: string };
    url: string;
    filename: string;
    size: number;
    primary?: boolean;
}

export interface ModrinthVersionDependency {
    dependency_type: string;
    project_id?: string;
}

export interface ModrinthVersion {
    id: string;
    files: ModrinthVersionFile[];
    game_versions: string[];
    loaders: string[];
    client_side: "required" | "optional" | "unsupported";
    server_side: "required" | "optional" | "unsupported";
    dependencies: ModrinthVersionDependency[];
}

export async function getVersionId(project: ModrinthProject, packData: Pack): Promise<string> {
    // Fetch all versions for this project
    const res = await fetch(`${BaseURL}/project/${project.id}/version`);
    if (!res.ok) throw new Error(`Failed to fetch versions for project ${project.slug}`);
    const versions: ModrinthVersion[] = await res.json();

    // Find the best matching version for the pack's gameVersion and modloader
    const matches = versions.filter((v: any) =>
        v.game_versions.includes(packData.gameVersion) &&
        v.loaders.includes(packData.modloader.name)
    );
    if (matches.length === 0) {
        throw new Error(`No compatible version found for ${project.slug} with game version ${packData.gameVersion} and loader ${packData.modloader.name}`);
    }
    // Prefer the latest version (assuming sorted by date desc)
    const version = matches[0];
    return version.id;
}

// Create a stub from a ModrinthProject and Pack
export async function projectToStub(project: ModrinthProject, packData: Pack) {
    const versionId = await getVersionId(project, packData);
    // Fetch version details
    const res = await fetch(`${BaseURL}/version/${versionId}`);
    if (!res.ok) throw new Error(`Failed to fetch version details for version ${versionId}`);
    const version: ModrinthVersion = await res.json();

    // Find the primary file (first file, or with primary: true)
    const file = version.files.find((f: any) => f.primary) || version.files[0];
    if (!file) throw new Error(`No files found for version ${versionId}`);
    if (project.project_type === "modpack") {
        throw new Error(`Project ${project.slug} is a modpack, which cannot be converted to a stub.`);
    }

    // Build the stub
    return new Stub({
        name: project.title,
        projectId: project.id,
        slug: project.slug,
        type: project.project_type,
        loader: packData.modloader.name,
        gameVersion: packData.gameVersion,
        hashes: {
            sha1: file.hashes.sha1,
            sha512: file.hashes.sha512,
        },
        download: {
            versionId: version.id,
            url: file.url,
            path: file.filename,
            size: file.size,
        },
        environments: {
            client: version.client_side,
            server: version.server_side,
        },
        dependencies: version.dependencies
            .filter((dep: any) => dep.dependency_type === "required" && dep.project_id)
            .map((dep: any) => dep.project_id),
    });
}