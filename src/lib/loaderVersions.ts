// shamelessly stolen from https://github.com/packwiz/packwiz/blob/main/core/versionutil.go, thanks

import { parseStringPromise } from 'xml2js';

type MavenMetadata = {
    metadata: {
        groupId: string[];
        artifactId: string[];
        versioning: [{
            release?: string[];
            latest?: string[];
            versions: [{
                version: string[];
            }];
            lastUpdated?: string[];
        }];
    };
};

type ModLoaderComponent = {
    name: string;
    friendlyName: string;
    versionListGetter: (mcVersion: string) => Promise<[string[], string]>;
};

export const ModLoaders: Record<string, ModLoaderComponent> = {
    fabric: {
        name: 'fabric',
        friendlyName: 'Fabric loader',
        versionListGetter: fetchMavenVersionList('https://maven.fabricmc.net/net/fabricmc/fabric-loader/maven-metadata.xml'),
    },
    forge: {
        name: 'forge',
        friendlyName: 'Forge',
        versionListGetter: fetchMavenVersionPrefixedListStrip('https://files.minecraftforge.net/maven/net/minecraftforge/forge/maven-metadata.xml', 'Forge'),
    },
    liteloader: {
        name: 'liteloader',
        friendlyName: 'LiteLoader',
        versionListGetter: fetchMavenVersionPrefixedList('https://repo.mumfrey.com/content/repositories/snapshots/com/mumfrey/liteloader/maven-metadata.xml', 'LiteLoader'),
    },
    quilt: {
        name: 'quilt',
        friendlyName: 'Quilt loader',
        versionListGetter: fetchMavenVersionList('https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-loader/maven-metadata.xml'),
    },
    neoforge: {
        name: 'neoforge',
        friendlyName: 'NeoForge',
        versionListGetter: fetchNeoForge(),
    },
};

function fetchMavenVersionList(url: string) {
    return async (_mcVersion: string): Promise<[string[], string]> => {
        const res = await fetch(url, { headers: { 'User-Agent': 'Mozilla/5.0', 'Accept': 'application/xml' } });
        const xml = await res.text();
        const out: MavenMetadata = await parseStringPromise(xml);
        const versions = out.metadata.versioning[0].versions[0].version;
        const release = out.metadata.versioning[0].release?.[0] || '';
        return [versions, release];
    };
}

function fetchMavenVersionFiltered(
    url: string,
    friendlyName: string,
    filter: (version: string, mcVersion: string) => boolean
) {
    return async (mcVersion: string): Promise<[string[], string]> => {
        const res = await fetch(url, { headers: { 'User-Agent': 'Mozilla/5.0', 'Accept': 'application/xml' } });
        const xml = await res.text();
        const out: MavenMetadata = await parseStringPromise(xml);
        const allVersions = out.metadata.versioning[0].versions[0].version;
        const allowedVersions = allVersions.filter(v => filter(v, mcVersion));
        if (allowedVersions.length === 0) throw new Error(`no ${friendlyName} versions available for this Minecraft version`);
        const release = out.metadata.versioning[0].release?.[0] || '';
        const latest = out.metadata.versioning[0].latest?.[0] || '';
        if (filter(release, mcVersion)) return [allowedVersions, release];
        if (filter(latest, mcVersion)) return [allowedVersions, latest];
        // Return the last (highest) version
        return [allowedVersions, allowedVersions[allowedVersions.length - 1]];
    };
}

function fetchMavenVersionPrefixedList(url: string, friendlyName: string) {
    return fetchMavenVersionFiltered(url, friendlyName, hasPrefixSplitDash);
}

function fetchMavenVersionPrefixedListStrip(url: string, friendlyName: string) {
    const noStrip = fetchMavenVersionPrefixedList(url, friendlyName);
    return async (mcVersion: string): Promise<[string[], string]> => {
        const [versions, latestVersion] = await noStrip(mcVersion);
        return [
            versions.map(v => removeMcVersion(v, mcVersion)),
            removeMcVersion(latestVersion, mcVersion),
        ];
    };
}

function removeMcVersion(str: string, mcVersion: string): string {
    return str.split('-').filter(v => v !== mcVersion).join('-');
}

function hasPrefixSplitDash(str: string, prefix: string): boolean {
    const components = str.split('-');
    return components.length > 1 && components[0] === prefix;
}

function fetchNeoForge() {
    const neoforgeOld = fetchMavenVersionPrefixedListStrip('https://maven.neoforged.net/releases/net/neoforged/forge/maven-metadata.xml', 'NeoForge');
    const neoforgeNew = fetchMavenWithNeoForgeStyleVersions('https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml', 'NeoForge');
    return async (mcVersion: string): Promise<[string[], string]> => {
        if (mcVersion === '1.20.1') {
            return neoforgeOld(mcVersion);
        } else {
            return neoforgeNew(mcVersion);
        }
    };
}

function fetchMavenWithNeoForgeStyleVersions(url: string, friendlyName: string) {
    return fetchMavenVersionFiltered(url, friendlyName, (neoforgeVersion, mcVersion) => {
        const mcSplit = mcVersion.split('.');
        if (mcSplit.length < 2) return false;
        const mcMajor = mcSplit[1];
        const mcMinor = mcSplit[2] || '0';
        return neoforgeVersion.startsWith(`${mcMajor}.${mcMinor}`);
    });
}

export function componentToFriendlyName(component: string): string {
    if (component === 'minecraft') return 'Minecraft';
    return ModLoaders[component]?.friendlyName || component;
}
