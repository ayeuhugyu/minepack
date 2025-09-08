package core

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/unascribed/FlexVer/go/flexver"
)

// MavenMetadata represents the structure of Maven metadata XML
type MavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Release  string `xml:"release"`
		Latest   string `xml:"latest"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
		LastUpdated string `xml:"lastUpdated"`
	} `xml:"versioning"`
}

// ModLoaderComponent represents a mod loader with its version fetching function
type ModLoaderComponent struct {
	Name              string
	FriendlyName      string
	VersionListGetter func(mcVersion string) ([]string, string, error)
}

// MinecraftVersion represents a Minecraft version from the manifest
type MinecraftVersion struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`
}

// MinecraftManifest represents the Minecraft version manifest
type MinecraftManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []MinecraftVersion `json:"versions"`
}

// ForgeRecommended represents the Forge promotions structure
type ForgeRecommended struct {
	Homepage string            `json:"homepage"`
	Versions map[string]string `json:"promos"`
}

// ModLoaders contains all supported mod loaders with their version fetchers
var ModLoaders = map[string]ModLoaderComponent{
	"fabric": {
		Name:              "fabric",
		FriendlyName:      "Fabric Loader",
		VersionListGetter: FetchMavenVersionList("https://maven.fabricmc.net/net/fabricmc/fabric-loader/maven-metadata.xml"),
	},
	"forge": {
		Name:              "forge",
		FriendlyName:      "Forge",
		VersionListGetter: FetchMavenVersionPrefixedListStrip("https://files.minecraftforge.net/maven/net/minecraftforge/forge/maven-metadata.xml", "Forge"),
	},
	"liteloader": {
		Name:              "liteloader",
		FriendlyName:      "LiteLoader",
		VersionListGetter: FetchMavenVersionPrefixedList("https://repo.mumfrey.com/content/repositories/snapshots/com/mumfrey/liteloader/maven-metadata.xml", "LiteLoader"),
	},
	"quilt": {
		Name:              "quilt",
		FriendlyName:      "Quilt Loader",
		VersionListGetter: FetchMavenVersionList("https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-loader/maven-metadata.xml"),
	},
	"neoforge": {
		Name:              "neoforge",
		FriendlyName:      "NeoForge",
		VersionListGetter: FetchNeoForge(),
	},
}

// GetWithUA makes an HTTP GET request with a user agent
func GetWithUA(url string, accept string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "minepack/1.0.0")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	return client.Do(req)
}

// FetchMinecraftVersions fetches all Minecraft versions from the official manifest
func FetchMinecraftVersions() (*MinecraftManifest, error) {
	res, err := GetWithUA("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var manifest MinecraftManifest
	if err := json.NewDecoder(res.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// GetLatestMinecraftVersion returns the latest release version of Minecraft
func GetLatestMinecraftVersion() (string, error) {
	manifest, err := FetchMinecraftVersions()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Release, nil
}

// GetLatestMinecraftSnapshot returns the latest snapshot version of Minecraft
func GetLatestMinecraftSnapshot() (string, error) {
	manifest, err := FetchMinecraftVersions()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Snapshot, nil
}

// FetchMavenVersionList creates a version fetcher for simple Maven repositories
func FetchMavenVersionList(url string) func(mcVersion string) ([]string, string, error) {
	return func(mcVersion string) ([]string, string, error) {
		res, err := GetWithUA(url, "application/xml")
		if err != nil {
			return []string{}, "", err
		}
		defer res.Body.Close()

		var metadata MavenMetadata
		if err := xml.NewDecoder(res.Body).Decode(&metadata); err != nil {
			return []string{}, "", err
		}

		return metadata.Versioning.Versions.Version, metadata.Versioning.Release, nil
	}
}

// FetchMavenVersionFiltered creates a version fetcher with custom filtering
func FetchMavenVersionFiltered(url string, friendlyName string, filter func(version string, mcVersion string) bool) func(mcVersion string) ([]string, string, error) {
	return func(mcVersion string) ([]string, string, error) {
		res, err := GetWithUA(url, "application/xml")
		if err != nil {
			return []string{}, "", err
		}
		defer res.Body.Close()

		var metadata MavenMetadata
		if err := xml.NewDecoder(res.Body).Decode(&metadata); err != nil {
			return []string{}, "", err
		}

		allowedVersions := make([]string, 0, len(metadata.Versioning.Versions.Version))
		for _, v := range metadata.Versioning.Versions.Version {
			if filter(v, mcVersion) {
				allowedVersions = append(allowedVersions, v)
			}
		}

		if len(allowedVersions) == 0 {
			return []string{}, "", errors.New("no " + friendlyName + " versions available for this Minecraft version")
		}

		if filter(metadata.Versioning.Release, mcVersion) {
			return allowedVersions, metadata.Versioning.Release, nil
		}
		if filter(metadata.Versioning.Latest, mcVersion) {
			return allowedVersions, metadata.Versioning.Latest, nil
		}

		// Sort list to get largest version
		flexver.VersionSlice(allowedVersions).Sort()
		return allowedVersions, allowedVersions[len(allowedVersions)-1], nil
	}
}

// FetchMavenVersionPrefixedList creates a version fetcher for versions with MC version prefix
func FetchMavenVersionPrefixedList(url string, friendlyName string) func(mcVersion string) ([]string, string, error) {
	return FetchMavenVersionFiltered(url, friendlyName, hasPrefixSplitDash)
}

// FetchMavenVersionPrefixedListStrip creates a version fetcher that strips MC version from results
func FetchMavenVersionPrefixedListStrip(url string, friendlyName string) func(mcVersion string) ([]string, string, error) {
	noStrip := FetchMavenVersionPrefixedList(url, friendlyName)
	return func(mcVersion string) ([]string, string, error) {
		versions, latestVersion, err := noStrip(mcVersion)
		if err != nil {
			return nil, "", err
		}

		for k, v := range versions {
			versions[k] = removeMcVersion(v, mcVersion)
		}
		latestVersion = removeMcVersion(latestVersion, mcVersion)
		return versions, latestVersion, nil
	}
}

// FetchNeoForge creates a version fetcher for NeoForge (handles both old and new versioning)
func FetchNeoForge() func(mcVersion string) ([]string, string, error) {
	// NeoForge changed versioning scheme for 1.20.2 and above
	neoforgeOld := FetchMavenVersionPrefixedListStrip("https://maven.neoforged.net/releases/net/neoforged/forge/maven-metadata.xml", "NeoForge")
	neoforgeNew := FetchMavenWithNeoForgeStyleVersions("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml", "NeoForge")

	return func(mcVersion string) ([]string, string, error) {
		if mcVersion == "1.20.1" {
			return neoforgeOld(mcVersion)
		} else {
			return neoforgeNew(mcVersion)
		}
	}
}

// FetchMavenWithNeoForgeStyleVersions creates a version fetcher for NeoForge's new versioning style
func FetchMavenWithNeoForgeStyleVersions(url string, friendlyName string) func(mcVersion string) ([]string, string, error) {
	return FetchMavenVersionFiltered(url, friendlyName, func(neoforgeVersion string, mcVersion string) bool {
		// Minecraft versions are in the form of 1.a.b
		// Neoforge versions are in the form of a.b.x
		// Eg, for minecraft 1.20.6, neoforge version 20.6.2 and 20.6.83-beta would both be valid versions
		mcSplit := strings.Split(mcVersion, ".")
		if len(mcSplit) < 2 {
			// This does not appear to be a minecraft version that's formatted in a way that matches neoforge
			return false
		}

		mcMajor := mcSplit[1]
		mcMinor := "0"
		if len(mcSplit) > 2 {
			mcMinor = mcSplit[2]
		}

		return strings.HasPrefix(neoforgeVersion, mcMajor+"."+mcMinor)
	})
}

// GetForgeRecommended gets the recommended version of Forge for the given Minecraft version
func GetForgeRecommended(mcVersion string) string {
	res, err := GetWithUA("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json", "application/json")
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	var recommended ForgeRecommended
	if err := json.NewDecoder(res.Body).Decode(&recommended); err != nil {
		return ""
	}

	// Get mcVersion-recommended, if it doesn't exist then get mcVersion-latest
	recommendedString := fmt.Sprintf("%s-recommended", mcVersion)
	if recommended.Versions[recommendedString] != "" {
		return recommended.Versions[recommendedString]
	}

	latestString := fmt.Sprintf("%s-latest", mcVersion)
	if recommended.Versions[latestString] != "" {
		return recommended.Versions[latestString]
	}

	return ""
}

// Helper functions

// removeMcVersion removes the Minecraft version from a version string
func removeMcVersion(str string, mcVersion string) string {
	components := strings.Split(str, "-")
	newComponents := make([]string, 0)
	for _, v := range components {
		if v != mcVersion {
			newComponents = append(newComponents, v)
		}
	}
	return strings.Join(newComponents, "-")
}

// hasPrefixSplitDash checks if a version string contains the given prefix when split by dashes
func hasPrefixSplitDash(str string, prefix string) bool {
	components := strings.Split(str, "-")
	if len(components) > 0 {
		return components[0] == prefix
	}
	return false
}

// ComponentToFriendlyName converts a component name to a user-friendly name
func ComponentToFriendlyName(component string) string {
	if component == "minecraft" {
		return "Minecraft"
	}
	loader, ok := ModLoaders[component]
	if ok {
		return loader.FriendlyName
	}
	return component
}

// GetModLoaderVersions returns all available versions and the recommended version for a mod loader
func GetModLoaderVersions(loaderName string, mcVersion string) ([]string, string, error) {
	loader, exists := ModLoaders[loaderName]
	if !exists {
		return nil, "", fmt.Errorf("unknown mod loader: %s", loaderName)
	}

	return loader.VersionListGetter(mcVersion)
}

// GetLatestModLoaderVersion returns just the latest/recommended version for a mod loader
func GetLatestModLoaderVersion(loaderName string, mcVersion string) (string, error) {
	_, latest, err := GetModLoaderVersions(loaderName, mcVersion)
	return latest, err
}

// GetAllLatestVersions returns the latest version of Minecraft and all mod loaders for a given MC version
func GetAllLatestVersions(mcVersion string) map[string]string {
	result := make(map[string]string)

	// Get latest Minecraft version if no specific version provided
	if mcVersion == "" {
		if latest, err := GetLatestMinecraftVersion(); err == nil {
			result["minecraft"] = latest
			mcVersion = latest
		}
	} else {
		result["minecraft"] = mcVersion
	}

	// Get latest versions for all mod loaders
	for name := range ModLoaders {
		if latest, err := GetLatestModLoaderVersion(name, mcVersion); err == nil {
			result[name] = latest
		} else {
			result[name] = "error: " + err.Error()
		}
	}

	return result
}
