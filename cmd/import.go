/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"minepack/core"
	mymodrinth "minepack/core/api/modrinth"
	"minepack/core/project"
	"minepack/util"
	"minepack/util/version"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

// ModrinthPack represents a Modrinth modpack format
type ModrinthPack struct {
	FormatVersion int    `json:"formatVersion"`
	Game          string `json:"game"`
	VersionID     string `json:"versionId"`
	Name          string `json:"name"`
	Summary       string `json:"summary"`
	Files         []struct {
		Path      string            `json:"path"`
		Hashes    map[string]string `json:"hashes"`
		Env       map[string]string `json:"env"`
		Downloads []string          `json:"downloads"`
		FileSize  int64             `json:"fileSize"`
	} `json:"files"`
	Dependencies map[string]string `json:"dependencies"`
}

// Progress message types for import operations
type hashingProgressMsg struct {
	current int
	total   int
	name    string
	err     error
}

type contentCreationProgressMsg struct {
	current int
	total   int
	err     error
}

type copyProgressMsg struct {
	current int
	total   int
	name    string
	err     error
}

// Progress models for different import operations
type importProgress struct {
	progress progress.Model
	current  int
	total    int
	stage    string
	err      error
}

func newImportProgress(total int, stage string) importProgress {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40
	return importProgress{
		progress: prog,
		total:    total,
		stage:    stage,
	}
}

func (m importProgress) Init() tea.Cmd {
	return nil
}

func (m importProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case hashingProgressMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.current = msg.current
		return m, nil
	case contentCreationProgressMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.current = msg.current
		return m, nil
	case copyProgressMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.current = msg.current
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m importProgress) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	percentage := float64(m.current) / float64(m.total)
	if m.total == 0 {
		percentage = 0
	}

	return fmt.Sprintf("%s %s (%d/%d)\n", m.stage, m.progress.ViewAs(percentage), m.current, m.total)
}

// ModrinthVersionFromHash represents response from version lookup by hash
type ModrinthVersionFromHash struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Files     []struct {
		Hashes struct {
			Sha1   string `json:"sha1"`
			Sha512 string `json:"sha512"`
		} `json:"hashes"`
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Primary  bool   `json:"primary"`
		Size     int64  `json:"size"`
	} `json:"files"`
}

// calculateSHA512 calculates SHA512 hash of a file
func calculateSHA512(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// bulkFetchProjects fetches multiple projects from Modrinth in a single API call
func bulkFetchProjects(projectIDs []string) (map[string]*modrinth.Project, error) {
	if len(projectIDs) == 0 {
		return make(map[string]*modrinth.Project), nil
	}

	// Build the URL with proper JSON array encoding for GET request
	baseURL := "https://api.modrinth.com/v2/projects"

	// Create JSON array string
	idsJSON, err := json.Marshal(projectIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IDs: %w", err)
	}

	// Create URL with proper encoding
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("ids", string(idsJSON))
	u.RawQuery = q.Encode()

	// Make GET request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Minepack/1.0 (+https://github.com/ayeuhugyu/minepack)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Read response body for debugging
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var projects []*modrinth.Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to map for easy lookup
	projectMap := make(map[string]*modrinth.Project)
	for _, proj := range projects {
		if proj.ID != nil {
			projectMap[*proj.ID] = proj
		}
	}

	return projectMap, nil
}

// copyDirectory recursively copies a directory and all its contents
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			// Copy file
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			return err
		}
	})
}

// importModrinthPack imports a Modrinth modpack
func importModrinthPack(packPath string) error {
	// Create temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "minepack-mrpack-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up when done

	fmt.Printf("extracting modpack archive...\n")

	// Open the .mrpack file as a ZIP archive
	zipReader, err := zip.OpenReader(packPath)
	if err != nil {
		return fmt.Errorf("failed to open .mrpack file as ZIP archive: %w", err)
	}
	defer zipReader.Close()

	// Extract all files from the ZIP archive
	for _, file := range zipReader.File {
		// Create the full file path
		destPath := filepath.Join(tempDir, file.Name)

		// Create directory if needed
		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.FileInfo().Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Name, err)
		}

		// Extract file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in archive: %w", file.Name, err)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	// Read the modrinth.index.json file
	indexPath := filepath.Join(tempDir, "modrinth.index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read modrinth.index.json: %w", err)
	}

	var pack ModrinthPack
	if err := json.Unmarshal(indexData, &pack); err != nil {
		return fmt.Errorf("failed to parse modrinth.index.json: %w", err)
	}

	fmt.Printf("importing modrinth pack: %s\n", pack.Name)
	fmt.Printf("description: %s\n", pack.Summary)
	fmt.Printf("game version: %s\n", pack.Dependencies["minecraft"])

	// Extract modloader info
	var modloader, modloaderVersion string
	if fabricVersion, ok := pack.Dependencies["fabric-loader"]; ok {
		modloader = "fabric"
		modloaderVersion = fabricVersion
	} else if forgeVersion, ok := pack.Dependencies["forge"]; ok {
		modloader = "forge"
		modloaderVersion = forgeVersion
	} else if quiltVersion, ok := pack.Dependencies["quilt-loader"]; ok {
		modloader = "quilt"
		modloaderVersion = quiltVersion
	} else if neoforgeVersion, ok := pack.Dependencies["neoforge"]; ok {
		modloader = "neoforge"
		modloaderVersion = neoforgeVersion
	}

	fmt.Printf("modloader: %s %s\n", modloader, modloaderVersion)
	fmt.Printf("files: %d\n\n", len(pack.Files))

	// Collect mod files (files in mods/ folder with .jar extension)
	var modFiles []struct {
		Path      string
		Hashes    map[string]string
		Downloads []string
		FileSize  int64
	}

	var overrideFiles []struct {
		Path      string
		Downloads []string
		FileSize  int64
	}

	for _, file := range pack.Files {
		if strings.HasPrefix(file.Path, "mods/") && strings.HasSuffix(file.Path, ".jar") {
			modFiles = append(modFiles, struct {
				Path      string
				Hashes    map[string]string
				Downloads []string
				FileSize  int64
			}{
				Path:      file.Path,
				Hashes:    file.Hashes,
				Downloads: file.Downloads,
				FileSize:  file.FileSize,
			})
		} else {
			// Non-mod files go to overrides
			overrideFiles = append(overrideFiles, struct {
				Path      string
				Downloads []string
				FileSize  int64
			}{
				Path:      file.Path,
				Downloads: file.Downloads,
				FileSize:  file.FileSize,
			})
		}
	}

	fmt.Printf("found %d mod files and %d override files\n", len(modFiles), len(overrideFiles))

	// Lookup mods on Modrinth using hashes
	if len(modFiles) > 0 {
		fmt.Println("looking up mods on modrinth...")

		var hashes []string
		hashToFile := make(map[string]struct {
			Path      string
			Downloads []string
			FileSize  int64
		})

		for _, file := range modFiles {
			if sha512Hash, ok := file.Hashes["sha512"]; ok {
				hashes = append(hashes, sha512Hash)
				hashToFile[sha512Hash] = struct {
					Path      string
					Downloads []string
					FileSize  int64
				}{
					Path:      file.Path,
					Downloads: file.Downloads,
					FileSize:  file.FileSize,
				}
			} else {
				// If no SHA512, try SHA1
				if sha1Hash, ok := file.Hashes["sha1"]; ok {
					hashes = append(hashes, sha1Hash)
					hashToFile[sha1Hash] = struct {
						Path      string
						Downloads []string
						FileSize  int64
					}{
						Path:      file.Path,
						Downloads: file.Downloads,
						FileSize:  file.FileSize,
					}
				}
			}
		}

		// Batch lookup all hashes
		var versions map[string]*modrinth.Version
		var lookupErr error

		err = spinner.New().
			Title("looking up mods on modrinth...").
			Type(spinner.Dots).
			Action(func() {
				versions, lookupErr = mymodrinth.ModrinthClient.VersionFiles.GetFromHashes(hashes, "sha512")
				if lookupErr != nil {
					// Try with SHA1 if SHA512 fails
					versions, lookupErr = mymodrinth.ModrinthClient.VersionFiles.GetFromHashes(hashes, "sha1")
				}
			}).
			Run()

		if err != nil {
			fmt.Printf(util.FormatError("spinner error: %v\n"), err)
			return fmt.Errorf("lookup failed: %w", err)
		}
		if lookupErr != nil {
			return fmt.Errorf("failed to lookup mods: %w", lookupErr)
		}

		// Collect project IDs for bulk lookup
		projectIDs := make([]string, 0, len(versions))
		versionToProject := make(map[string]string)

		for hash, version := range versions {
			if version.ProjectID != nil {
				projectID := *version.ProjectID
				projectIDs = append(projectIDs, projectID)
				versionToProject[hash] = projectID
			}
		}

		// Bulk fetch project data
		var projects map[string]*modrinth.Project
		var fetchErr error

		err = spinner.New().
			Title("fetching project details...").
			Type(spinner.Dots).
			Action(func() {
				projects, fetchErr = bulkFetchProjects(projectIDs)
			}).
			Run()

		if err != nil {
			fmt.Printf(util.FormatError("spinner error: %v\n"), err)
			return fmt.Errorf("fetch failed: %w", err)
		}
		if fetchErr != nil {
			return fmt.Errorf("failed to fetch project details: %w", fetchErr)
		}

		var foundMods []string
		var notFoundFiles []struct {
			Path      string
			Downloads []string
			FileSize  int64
		}

		for _, hash := range hashes {
			fileInfo := hashToFile[hash]
			if version, found := versions[hash]; found {
				if version.ProjectID != nil {
					foundMods = append(foundMods, *version.ProjectID)
				}
			} else {
				notFoundFiles = append(notFoundFiles, fileInfo)
			}
		}

		fmt.Printf("\nanalysis complete:\n")
		fmt.Printf("- found on modrinth: %d\n", len(foundMods))
		fmt.Printf("- not found: %d\n", len(notFoundFiles))

		if len(foundMods) == 0 && len(notFoundFiles) == 0 {
			fmt.Println(util.FormatError("no mods found to import"))
			return nil
		}

		// Ask for project details
		var name, description, author string

		if pack.Name != "" {
			name = pack.Name
		}
		if pack.Summary != "" {
			description = pack.Summary
		}

		metaForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("project name").
					Description("name for the imported project").
					Placeholder("imported-pack").
					Value(&name),
				huh.NewText().
					Title("description").
					Description("description for the imported project").
					Lines(3).
					Placeholder("imported from modrinth pack").
					Value(&description),
				huh.NewInput().
					Title("author").
					Description("author of the project").
					Placeholder("unknown").
					Value(&author),
			),
		)

		if err := metaForm.Run(); err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if name == "" {
			name = "imported-pack"
		}
		if description == "" {
			description = "imported from modrinth pack"
		}
		if author == "" {
			author = "unknown"
		}

		// Create project structure
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		projectData := project.Project{
			Name:          name,
			Description:   description,
			Author:        author,
			DefaultSource: "modrinth",
			Root:          filepath.Join(cwd, strings.ReplaceAll(name, " ", "-")),
			Versions: project.ProjectVersions{
				Game: pack.Dependencies["minecraft"],
				Loader: project.ModloaderVersion{
					Name:    modloader,
					Version: modloaderVersion,
				},
				Minepack: version.Version,
			},
		}

		// Write project files
		if err := project.WriteProject(&projectData); err != nil {
			return fmt.Errorf("failed to write project files: %w", err)
		}

		fmt.Printf("\nproject created at: %s\n", projectData.Root)

		// Create content files for found mods
		if len(foundMods) > 0 {
			var contentCreationErrors []string
			var createdCount int

			prog := newImportProgress(len(hashes), "creating content files...")
			p := tea.NewProgram(prog)

			go func() {
				current := 0
				for _, hash := range hashes {
					if version, found := versions[hash]; found && version.ProjectID != nil {
						projectID := *version.ProjectID

						// Get project data for content creation
						if project, exists := projects[projectID]; exists {
							// Convert project to content data
							contentData := mymodrinth.ConvertProjectToContentData(project, version)

							// Add content to project (this updates content.mp.sum.yaml)
							if err := projectData.AddContent(contentData); err != nil {
								projectTitle := projectID
								if project.Title != nil {
									projectTitle = *project.Title
								}
								contentCreationErrors = append(contentCreationErrors, fmt.Sprintf("failed to add content for %s: %v", projectTitle, err))
							} else {
								createdCount++
							}
						}
					}
					current++
					p.Send(contentCreationProgressMsg{current: current, total: len(hashes)})
				}
				p.Quit()
			}()

			if _, err := p.Run(); err != nil {
				fmt.Printf(util.FormatError("content creation failed: %v\n"), err)
			} else {
				fmt.Printf("created %d content files\n", createdCount)
			}

			// Report any content creation errors
			for _, errMsg := range contentCreationErrors {
				fmt.Println(util.FormatWarning(errMsg))
			}
		}

		// Handle not found mods and other files as overrides
		if len(notFoundFiles) > 0 || len(overrideFiles) > 0 {
			overridesPath := filepath.Join(projectData.Root, "overrides")
			if err := os.MkdirAll(overridesPath, os.ModePerm); err != nil {
				fmt.Printf(util.FormatWarning("failed to create overrides directory: %v\n"), err)
			}
		}

		// Copy overrides folder from extracted archive if it exists
		archiveOverridesPath := filepath.Join(tempDir, "overrides")
		if _, err := os.Stat(archiveOverridesPath); err == nil {
			projectOverridesPath := filepath.Join(projectData.Root, "overrides")

			err = spinner.New().
				Title("copying overrides from archive...").
				Type(spinner.Dots).
				Action(func() {
					err = copyDirectory(archiveOverridesPath, projectOverridesPath)
				}).
				Run()

			if err != nil {
				fmt.Printf(util.FormatWarning("failed to copy overrides from archive: %v\n"), err)
			} else {
				fmt.Println("copied overrides from archive")
			}
		}

		fmt.Printf("\nsuccessfully imported modpack!\n")
		fmt.Printf("- modrinth mods: %d\n", len(foundMods))
		if len(notFoundFiles) > 0 {
			fmt.Printf("- override mods: %d\n", len(notFoundFiles))
		}
		if len(overrideFiles) > 0 {
			fmt.Printf("- override files: %d\n", len(overrideFiles))
		}
	} else {
		fmt.Println("no mod files found in pack")
	}

	return nil
}

// importMinecraftInstance imports a Minecraft instance
func importMinecraftInstance(instancePath string) error {
	// Check if it's a valid minecraft instance
	modsPath := filepath.Join(instancePath, "mods")
	if _, err := os.Stat(modsPath); os.IsNotExist(err) {
		fmt.Println(util.FormatError("not a valid minecraft instance: mods folder not found"))
		return nil
	}

	// Get list of mod files
	modFiles, err := filepath.Glob(filepath.Join(modsPath, "*.jar"))
	if err != nil {
		fmt.Printf(util.FormatError("failed to list mod files: %w"), err)
		return nil
	}

	if len(modFiles) == 0 {
		fmt.Println(util.FormatError("no mod files found in instance"))
		return nil
	}

	fmt.Printf("found %d mod files in instance\n", len(modFiles))

	// Ask for version and modloader info (reuse logic from init command)
	var gameVersion, selectedModloader string

	// Fetch minecraft versions for validation
	var allGameVersions *core.MinecraftManifest
	var fetchErr error

	err = spinner.New().
		Title("fetching minecraft versions...").
		Type(spinner.Dots).
		Action(func() {
			allGameVersions, fetchErr = core.FetchMinecraftVersions()
		}).
		Run()

	if err != nil {
		fmt.Println(util.FormatError("spinner error: %w"), err)
		return nil
	}
	if fetchErr != nil {
		fmt.Println(util.FormatError("failed to fetch minecraft versions: %w"), fetchErr)
		return nil
	}

	var allGameVersionsFlat []string
	for _, v := range allGameVersions.Versions {
		allGameVersionsFlat = append(allGameVersionsFlat, v.ID)
	}

	// Get game version
	var inputGameVersion string
	versionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("game version").
				Description("enter minecraft version for this instance").
				Placeholder("1.20.1").
				Suggestions(allGameVersionsFlat).
				Value(&inputGameVersion),
		),
	)

	if err := versionForm.Run(); err != nil {
		fmt.Println(util.FormatError("prompt failed: %w"), err)
		return nil
	}

	gameVersion = inputGameVersion
	if gameVersion == "" {
		gameVersion = "1.20.1"
	}

	// Fetch modloader versions
	var allModloaderVersions map[string]string

	err = spinner.New().
		Title("fetching modloader versions...").
		Type(spinner.Dots).
		Action(func() {
			allModloaderVersions = core.GetAllLatestVersions(gameVersion)
		}).
		Run()

	if err != nil {
		fmt.Println(util.FormatError("spinner error: %w"), err)
		return nil
	}

	// Get modloader
	modloaderOrder := []string{"fabric", "forge", "quilt", "neoforge", "liteloader"}
	var availableModloaderNames []string

	for _, name := range modloaderOrder {
		if version, exists := allModloaderVersions[name]; exists {
			if name == "minecraft" {
				continue
			}
			if strings.HasPrefix(version, "error:") {
				continue
			}
			availableModloaderNames = append(availableModloaderNames, name)
		}
	}

	modloaderForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("modloader").
				Description("choose the modloader used by this instance").
				Options(huh.NewOptions(availableModloaderNames...)...).
				Value(&selectedModloader),
		),
	)

	if err := modloaderForm.Run(); err != nil {
		fmt.Println(util.FormatError("prompt failed: %w"), err)
		return nil
	}

	if selectedModloader == "" {
		selectedModloader = "fabric"
	}

	fmt.Printf("\ngame version: %s\n", gameVersion)
	fmt.Printf("modloader: %s %s\n\n", selectedModloader, allModloaderVersions[selectedModloader])

	// Hash mods and lookup on Modrinth
	fmt.Println("analyzing mods...")

	// First, hash all mods with progress bar
	var hashes []string
	hashToFile := make(map[string]string)
	var hashingErrors []string

	if len(modFiles) > 0 {
		prog := newImportProgress(len(modFiles), "hashing mod files...")
		p := tea.NewProgram(prog)

		go func() {
			for i, modFile := range modFiles {
				filename := filepath.Base(modFile)

				hash, err := calculateSHA512(modFile)
				if err != nil {
					hashingErrors = append(hashingErrors, fmt.Sprintf("failed to hash %s: %s", filename, err))
					p.Send(hashingProgressMsg{current: i + 1, total: len(modFiles), err: err})
					continue
				}

				hashes = append(hashes, hash)
				hashToFile[hash] = filename
				p.Send(hashingProgressMsg{current: i + 1, total: len(modFiles), name: filename})
			}
			p.Quit()
		}()

		if _, err := p.Run(); err != nil {
			fmt.Printf(util.FormatError("hashing failed: %v\n"), err)
			return nil
		}
	}

	// Report any hashing errors
	for _, errMsg := range hashingErrors {
		fmt.Println(util.FormatWarning(errMsg))
	}

	// Batch lookup all hashes
	fmt.Printf("\nlooking up %d mods on modrinth...\n", len(hashes))

	var versions map[string]*modrinth.Version
	var lookupErr error

	err = spinner.New().
		Title("looking up mods on modrinth...").
		Type(spinner.Dots).
		Action(func() {
			versions, lookupErr = mymodrinth.ModrinthClient.VersionFiles.GetFromHashes(hashes, "sha512")
		}).
		Run()

	if err != nil {
		fmt.Printf(util.FormatError("spinner error: %v\n"), err)
		return nil
	}
	if lookupErr != nil {
		fmt.Printf(util.FormatError("failed to lookup mods: %v\n"), lookupErr)
		return nil
	}

	if len(versions) == 0 {
		fmt.Println(util.FormatError("no mods found on modrinth"))
		return nil
	}

	// Collect project IDs for bulk lookup
	projectIDs := make([]string, 0, len(versions))
	versionToProject := make(map[string]string)

	for hash, version := range versions {
		if version.ProjectID != nil {
			projectID := *version.ProjectID
			projectIDs = append(projectIDs, projectID)
			versionToProject[hash] = projectID
		}
	}

	// Bulk fetch project data
	var projects map[string]*modrinth.Project
	var projectFetchErr error

	err = spinner.New().
		Title("fetching project details...").
		Type(spinner.Dots).
		Action(func() {
			projects, projectFetchErr = bulkFetchProjects(projectIDs)
		}).
		Run()

	if err != nil {
		fmt.Printf(util.FormatError("spinner error: %v\n"), err)
		return nil
	}
	if projectFetchErr != nil {
		return fmt.Errorf("failed to fetch project details: %w", projectFetchErr)
	}

	var foundMods []string
	var notFoundMods []string

	for _, hash := range hashes {
		filename := hashToFile[hash]
		if version, found := versions[hash]; found {
			if version.ProjectID != nil {
				foundMods = append(foundMods, *version.ProjectID)
			}
		} else {
			notFoundMods = append(notFoundMods, filename)
		}
	}

	fmt.Printf("\nanalysis complete:\n")
	fmt.Printf("- found on modrinth: %d\n", len(foundMods))
	fmt.Printf("- not found: %d\n", len(notFoundMods))

	if len(notFoundMods) > 0 {
		fmt.Printf("\nmods not found on modrinth:\n")
		for _, mod := range notFoundMods {
			fmt.Printf("  - %s\n", mod)
		}
	}

	if len(foundMods) == 0 {
		fmt.Println(util.FormatError("no mods found on modrinth to import"))
		return nil
	}

	// Ask for project details
	var name, description, author string

	metaForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("project name").
				Description("name for the imported project").
				Placeholder("imported-pack").
				Value(&name),
			huh.NewText().
				Title("description").
				Description("description for the imported project").
				Lines(3).
				Placeholder("imported from minecraft instance").
				Value(&description),
			huh.NewInput().
				Title("author").
				Description("author of the project").
				Placeholder("unknown").
				Value(&author),
		),
	)

	if err := metaForm.Run(); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if name == "" {
		name = "imported-pack"
	}
	if description == "" {
		description = "imported from minecraft instance"
	}
	if author == "" {
		author = "unknown"
	}

	// Create project structure
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectData := project.Project{
		Name:          name,
		Description:   description,
		Author:        author,
		DefaultSource: "modrinth",
		Root:          filepath.Join(cwd, strings.ReplaceAll(name, " ", "-")),
		Versions: project.ProjectVersions{
			Game: gameVersion,
			Loader: project.ModloaderVersion{
				Name:    selectedModloader,
				Version: allModloaderVersions[selectedModloader],
			},
			Minepack: version.Version,
		},
	}

	// Write project files
	if err := project.WriteProject(&projectData); err != nil {
		return fmt.Errorf("failed to write project files: %w", err)
	}

	fmt.Printf("\nproject created at: %s\n", projectData.Root)

	// Create content files for found mods
	var contentCreationErrors []string
	var createdCount int

	if len(foundMods) > 0 {
		prog := newImportProgress(len(hashes), "Creating content files...")
		p := tea.NewProgram(prog)

		go func() {
			current := 0
			for _, hash := range hashes {
				if version, found := versions[hash]; found && version.ProjectID != nil {
					projectID := *version.ProjectID

					// Get project data for content creation
					if project, exists := projects[projectID]; exists {
						// Convert project to content data
						contentData := mymodrinth.ConvertProjectToContentData(project, version)

						// Add content to project (this updates content.mp.sum.yaml)
						if err := projectData.AddContent(contentData); err != nil {
							projectTitle := projectID
							if project.Title != nil {
								projectTitle = *project.Title
							}
							contentCreationErrors = append(contentCreationErrors, fmt.Sprintf("failed to add content for %s: %v", projectTitle, err))
						} else {
							createdCount++
						}
					}
				}
				current++
				p.Send(contentCreationProgressMsg{current: current, total: len(hashes)})
			}
			p.Quit()
		}()

		if _, err := p.Run(); err != nil {
			fmt.Printf(util.FormatError("content creation failed: %v\n"), err)
			return nil
		}

		fmt.Printf("created %d content files\n", createdCount)

		// Report any content creation errors
		for _, errMsg := range contentCreationErrors {
			fmt.Println(util.FormatWarning(errMsg))
		}
	}

	// Copy unfound mods to overrides folder
	if len(notFoundMods) > 0 {
		fmt.Printf("\ncopying %d unfound mods to overrides...\n", len(notFoundMods))

		overridesModsPath := filepath.Join(projectData.Root, "overrides", "mods")
		if err := os.MkdirAll(overridesModsPath, os.ModePerm); err != nil {
			fmt.Printf(util.FormatWarning("failed to create overrides/mods directory: %v\n"), err)
		} else {
			var copyErrors []string
			var copiedCount int

			prog := newImportProgress(len(notFoundMods), "copying mods to overrides...")
			p := tea.NewProgram(prog)

			go func() {
				for i, filename := range notFoundMods {
					// Find the original file path
					var originalPath string
					for _, modFile := range modFiles {
						if filepath.Base(modFile) == filename {
							originalPath = modFile
							break
						}
					}

					if originalPath != "" {
						destPath := filepath.Join(overridesModsPath, filename)

						// Copy the file
						sourceFile, err := os.Open(originalPath)
						if err != nil {
							copyErrors = append(copyErrors, fmt.Sprintf("failed to open %s: %v", filename, err))
							p.Send(copyProgressMsg{current: i + 1, total: len(notFoundMods), err: err})
							continue
						}

						destFile, err := os.Create(destPath)
						if err != nil {
							sourceFile.Close()
							copyErrors = append(copyErrors, fmt.Sprintf("failed to create %s: %v", destPath, err))
							p.Send(copyProgressMsg{current: i + 1, total: len(notFoundMods), err: err})
							continue
						}

						_, err = io.Copy(destFile, sourceFile)
						sourceFile.Close()
						destFile.Close()

						if err != nil {
							copyErrors = append(copyErrors, fmt.Sprintf("failed to copy %s: %v", filename, err))
						} else {
							copiedCount++
						}
					}
					p.Send(copyProgressMsg{current: i + 1, total: len(notFoundMods), name: filename})
				}
				p.Quit()
			}()

			if _, err := p.Run(); err != nil {
				fmt.Printf(util.FormatError("copying failed: %v\n"), err)
			} else {
				fmt.Printf("copied %d mods to overrides\n", copiedCount)
			}

			// Report any copy errors
			for _, errMsg := range copyErrors {
				fmt.Println(util.FormatWarning(errMsg))
			}
		}
	}

	fmt.Printf("\nsuccessfully imported project with %d mods!\n", len(foundMods))
	if len(notFoundMods) > 0 {
		fmt.Printf("+ %d mods copied to overrides (not found on Modrinth)\n", len(notFoundMods))
	}

	// Automatically link the project to the source instance
	fmt.Println("\nlinking project to source instance...")
	absInstancePath, err := filepath.Abs(instancePath)
	if err != nil {
		fmt.Printf(util.FormatWarning("warning: failed to get absolute path for linking: %v\n"), err)
	} else {
		// Load existing linked folders
		linked, err := loadLinkedFolders(projectData.Root)
		if err != nil {
			fmt.Printf(util.FormatWarning("warning: failed to load linked folders: %v\n"), err)
		} else {
			// Check if already linked (shouldn't be, but just in case)
			alreadyLinked := false
			for _, existingPath := range linked.Links {
				if existingPath == absInstancePath {
					alreadyLinked = true
					break
				}
			}

			if !alreadyLinked {
				// Add the new link
				linked.Links = append(linked.Links, absInstancePath)

				// Save the updated list
				if err := saveLinkedFolders(projectData.Root, linked); err != nil {
					fmt.Printf(util.FormatWarning("warning: failed to save link: %v\n"), err)
				} else {
					fmt.Printf(util.FormatSuccess("linked to instance: %s\n"), absInstancePath)
				}
			}
		}
	}

	// Helpful message about getting overrides
	fmt.Printf("\nrun 'minepack link get-overrides' to copy any custom files\n")
	fmt.Printf("   (configs, resource packs, etc.) from the instance to your project\n")

	return nil
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "import a modpack from various formats",
	Long:  `import a modpack from modrinth pack format or minecraft instance folder`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importPath := args[0]

		// Check if path exists
		if _, err := os.Stat(importPath); os.IsNotExist(err) {
			fmt.Printf(util.FormatError("path does not exist: %s\n"), importPath)
			return
		}

		// Check if it's a file or directory
		info, err := os.Stat(importPath)
		if err != nil {
			fmt.Printf(util.FormatError("failed to check path: %s\n"), err)
			return
		}

		if info.IsDir() {
			// Assume it's a minecraft instance
			fmt.Println("importing minecraft instance...")
			if err := importMinecraftInstance(importPath); err != nil {
				fmt.Printf(util.FormatError("failed to import instance: %s\n"), err)
				return
			}
		} else {
			// Assume it's a modrinth pack file
			if !strings.HasSuffix(strings.ToLower(importPath), ".mrpack") {
				fmt.Print(util.FormatWarning("warning: file doesn't have .mrpack extension\n"))
			}

			fmt.Println("importing modrinth pack...")
			if err := importModrinthPack(importPath); err != nil {
				fmt.Printf(util.FormatError("failed to import pack: %s\n"), err)
				return
			}
		}

		fmt.Print(util.FormatSuccess("import completed!"))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
