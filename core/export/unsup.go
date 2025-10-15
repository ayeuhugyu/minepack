package export

import (
	"encoding/json"
	"fmt"
	"minepack/core/project"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// UnsupManifest represents the unsup manifest format
type UnsupManifest struct {
	Name         string               `json:"name"`
	Versions     UnsupVersions        `json:"versions"`
	Creator      *UnsupCreator        `json:"creator,omitempty"`
	Flavors      []UnsupFlavor        `json:"flavors,omitempty"`
	UnsupVersion string               `json:"unsup_manifest"`
	Files        map[string]UnsupFile `json:"files,omitempty"`
}

type UnsupVersions struct {
	Current UnsupVersion   `json:"current"`
	History []UnsupVersion `json:"history"`
}

type UnsupVersion struct {
	Name string `json:"name"`
	Code int    `json:"code"`
}

type UnsupCreator struct {
	Ignore []string `json:"ignore,omitempty"`
}

type UnsupFlavor struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Envs []string `json:"envs"`
}

type UnsupFile struct {
	URL    string            `json:"url"`
	Hash   string            `json:"hash"`
	Method string            `json:"method,omitempty"`
	Size   int64             `json:"size,omitempty"`
	Meta   map[string]string `json:"meta,omitempty"`
}

// UnsupData represents user-defined unsup configuration (placeholder for future unsup-data.mp.yaml)
type UnsupData struct {
	Flavors []UnsupFlavor `yaml:"flavors,omitempty"`
	Ignore  []string      `yaml:"ignore,omitempty"`
}

// Convert minepack project to unsup manifest
func convertToUnsup(projectPath string) (*UnsupManifest, error) {
	// Parse minepack project
	packData, err := project.ParseProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	// Read unsup-data.mp.yaml if it exists (for future use)
	unsupData := &UnsupData{}
	unsupDataFile := filepath.Join(projectPath, "unsup-data.mp.yaml")
	if _, err := os.Stat(unsupDataFile); err == nil {
		unsupDataContent, err := os.ReadFile(unsupDataFile)
		if err == nil {
			yaml.Unmarshal(unsupDataContent, unsupData) // Ignore errors for now
		}
	}

	// Generate version code from version string
	versionCode := generateVersionCode("v0.0.0") // placeholder version, eventually i'd like to create an actual like "version add" command or something to do that with (see #17)

	// Create default ignore list
	ignoreList := []string{
		"saves",
		"screenshots",
		"logs",
		"crash-reports",
		"*.log",
		"options.txt",
		"servers.dat",
		"usernamecache.json",
	}

	// Merge with user-defined ignore patterns
	if len(unsupData.Ignore) > 0 {
		ignoreList = append(ignoreList, unsupData.Ignore...)
	}

	// Create unsup manifest
	manifest := &UnsupManifest{
		Name: packData.Name,
		Versions: UnsupVersions{
			Current: UnsupVersion{
				Name: "v0.0.0",
				Code: versionCode,
			},
			History: []UnsupVersion{
				{
					Name: "v0.0.0",
					Code: versionCode,
				},
			},
		},
		Creator: &UnsupCreator{
			Ignore: ignoreList,
		},
		UnsupVersion: "root-1",
		Files:        make(map[string]UnsupFile),
	}

	// Use user-defined flavors if available, otherwise create default
	if len(unsupData.Flavors) > 0 {
		manifest.Flavors = unsupData.Flavors
	} else if packData.Versions.Loader.Name != "" {
		// Add default client flavor based on loader
		manifest.Flavors = []UnsupFlavor{
			{
				ID:   "client",
				Name: "Client",
				Envs: []string{"client"},
			},
		}
	} // Get all content from project
	allContent, err := packData.GetAllContent()
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// Add content to manifest
	if err := addContentToUnsupManifest(manifest, allContent); err != nil {
		return nil, fmt.Errorf("failed to add content: %w", err)
	}

	return manifest, nil
}

func addContentToUnsupManifest(manifest *UnsupManifest, allContent []project.ContentData) error {
	// Process each content item
	for _, content := range allContent {
		// Skip if no download URL available
		if content.DownloadUrl == "" {
			continue // TODO: Add warning about missing download URL
		}

		// Create unsup file entry
		var hash, method string
		if content.File.Hashes.Sha256 != "" {
			hash = content.File.Hashes.Sha256
			method = "sha256"
		} else if content.File.Hashes.Sha1 != "" {
			hash = content.File.Hashes.Sha1
			method = "sha1"
		}

		// Determine file path based on content type
		var folder string
		switch content.ContentType {
		case project.Mod:
			folder = "mods"
		case project.Resourcepack:
			folder = "resourcepacks"
		case project.Shaderpack:
			folder = "shaderpacks"
		default:
			folder = "mods" // default to mods
		}

		filePath := folder + "/" + content.File.Filename
		manifest.Files[filePath] = UnsupFile{
			URL:    content.DownloadUrl,
			Hash:   hash,
			Method: method,
			Size:   content.File.Filesize,
			Meta: map[string]string{
				"source":   sourceToString(content.Source),
				"slug":     content.Slug,
				"mod_id":   content.Id,
				"mod_name": content.Name,
			},
		}
	}

	return nil
}

// sourceToString converts project.Source to string
func sourceToString(source project.Source) string {
	switch source {
	case project.Modrinth:
		return "modrinth"
	case project.Curseforge:
		return "curseforge"
	case project.Custom:
		return "custom"
	default:
		return "unknown"
	}
}

// Generate a numeric version code from semver string
func generateVersionCode(version string) int {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Split by dots and convert to numbers
	parts := strings.Split(version, ".")
	code := 0
	multiplier := 10000

	for i, part := range parts {
		if i >= 3 { // Only process major.minor.patch
			break
		}

		num, err := strconv.Atoi(part)
		if err != nil {
			num = 0
		}

		code += num * multiplier
		multiplier /= 100
	}

	return code
}

// Export example - this would be called from a CLI command
func ExportUnsup(projectPath, outputPath string) error {
	manifest, err := convertToUnsup(projectPath)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Write unsup-manifest.json
	manifestFile := filepath.Join(outputPath, "unsup-manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestFile, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Write unsup.ini config
	iniContent := fmt.Sprintf(`[unsup]
source = unsup-manifest.json
name = %s
created = %s
`, manifest.Name, time.Now().Format("2006-01-02 15:04:05"))

	iniFile := filepath.Join(outputPath, "unsup.ini")
	if err := os.WriteFile(iniFile, []byte(iniContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Unsup manifest exported to: %s\n", outputPath)
	fmt.Printf("- unsup-manifest.json: %d files\n", len(manifest.Files))
	fmt.Printf("- unsup.ini: configuration file\n")

	return nil
}
