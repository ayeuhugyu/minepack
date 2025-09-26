/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"minepack/core/project"
	"minepack/util"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

// formatFileSize formats bytes into human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getMissingFiles returns a list of files that are missing from the given link directory
func getMissingFiles(linkPath string, allContent []project.ContentData) ([]project.ContentData, error) {
	var missingFiles []project.ContentData

	for _, content := range allContent {
		// Check if the file exists in the link's mods directory
		modPath := filepath.Join(linkPath, content.File.Filepath)
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, content)
		}
	}

	return missingFiles, nil
}

// downloadFile downloads a file from URL to the specified path with progress tracking
func downloadFileWithProgress(url, filepath string, workerID int, name string, progressCh chan<- downloadMsg) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create progress writer
	pw := &progressWriter{
		workerID:   workerID,
		name:       name,
		total:      resp.ContentLength,
		progressCh: progressCh,
	}

	// Write the body to file with progress tracking
	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			return copyFile(path, dstPath)
		}
	})
}

var updateSummaryStyle = lipgloss.NewStyle().
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Margin(0, 1)

// Progress bar model for downloads
type downloadProgress struct {
	globalProgress   progress.Model
	workerProgresses [3]progress.Model
	workerNames      [3]string
	workerActive     [3]bool
	workerProgress   [3]float64
	current          int
	total            int
	done             bool
	err              error
}

func newDownloadProgress(total int) downloadProgress {
	globalProg := progress.New(progress.WithDefaultGradient())
	globalProg.Width = 60

	var workerProgs [3]progress.Model
	for i := 0; i < 3; i++ {
		workerProgs[i] = progress.New(progress.WithDefaultGradient())
		workerProgs[i].Width = 40
	}

	return downloadProgress{
		globalProgress:   globalProg,
		workerProgresses: workerProgs,
		total:            total,
	}
}

type downloadMsg struct {
	workerID int
	name     string
	progress float64 // 0.0 to 1.0 for real-time progress
	err      error
	complete bool
}

type downloadCompleteMsg struct{}

// progressWriter tracks download progress and sends updates
type progressWriter struct {
	workerID   int
	name       string
	total      int64
	downloaded int64
	progressCh chan<- downloadMsg
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	if pw.total > 0 {
		progress := float64(pw.downloaded) / float64(pw.total)
		// Send progress update
		select {
		case pw.progressCh <- downloadMsg{
			workerID: pw.workerID,
			name:     pw.name,
			progress: progress,
			complete: false,
		}:
		default:
			// Don't block if channel is full
		}
	}

	return n, nil
}

func (m downloadProgress) Init() tea.Cmd {
	return nil
}

func (m downloadProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case downloadMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		// Update worker progress
		if msg.workerID >= 0 && msg.workerID < 3 {
			if msg.complete {
				m.workerActive[msg.workerID] = false
				m.workerNames[msg.workerID] = ""
				m.workerProgress[msg.workerID] = 1.0
				m.current++
			} else {
				m.workerActive[msg.workerID] = true
				m.workerNames[msg.workerID] = msg.name
				m.workerProgress[msg.workerID] = msg.progress
			}
		}

		if m.current >= m.total {
			m.done = true
			return m, tea.Quit
		}
		return m, nil
	case downloadCompleteMsg:
		m.done = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		globalModel, cmd := m.globalProgress.Update(msg)
		m.globalProgress = globalModel.(progress.Model)

		for i := 0; i < 3; i++ {
			workerModel, _ := m.workerProgresses[i].Update(msg)
			m.workerProgresses[i] = workerModel.(progress.Model)
		}
		return m, cmd
	}
}

func (m downloadProgress) View() string {
	if m.err != nil {
		return fmt.Sprintf("download failed: %s\n", m.err.Error())
	}

	if m.done {
		return fmt.Sprintf(util.FormatSuccess("downloaded all %d files\n"), m.total)
	}

	// Global progress
	globalPercent := float64(m.current) / float64(m.total)
	view := fmt.Sprintf("overall progress: %d/%d\n%s\n\n",
		m.current, m.total, m.globalProgress.ViewAs(globalPercent))

	// Individual worker progress bars
	view += "downloading:\n"
	for i := 0; i < 3; i++ {
		if m.workerActive[i] {
			view += fmt.Sprintf("worker %d: %s\n%s\n",
				i+1, m.workerNames[i], m.workerProgresses[i].ViewAs(m.workerProgress[i]))
		} else {
			view += fmt.Sprintf("worker %d: waiting...\n%s\n",
				i+1, m.workerProgresses[i].ViewAs(m.workerProgress[i]))
		}
	}

	return view
}

// downloadWorker handles downloading files in parallel
func downloadWorker(workerID int, jobs <-chan project.ContentData, results chan<- downloadMsg, cacheDir string) {
	for content := range jobs {
		// Signal start of download
		results <- downloadMsg{workerID: workerID, name: content.Name, progress: 0.0, complete: false}

		cachePath := filepath.Join(cacheDir, content.File.Filepath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
			results <- downloadMsg{workerID: workerID, name: content.Name, err: fmt.Errorf("failed to create cache directory: %w", err), complete: true}
			continue
		}

		// Skip if already exists in cache
		if _, err := os.Stat(cachePath); err == nil {
			results <- downloadMsg{workerID: workerID, name: content.Name, progress: 1.0, complete: true}
			continue
		}

		err := downloadFileWithProgress(content.DownloadUrl, cachePath, workerID, content.Name, results)
		results <- downloadMsg{workerID: workerID, name: content.Name, progress: 1.0, err: err, complete: true}
	}
}

// filterContent filters content based on flags with dependency resolution
func filterContent(allContent []project.ContentData, serverOnly, clientOnly bool, source string) []project.ContentData {
	var filtered []project.ContentData

	// First pass: build a map for quick lookup
	contentMap := make(map[string]project.ContentData)
	for _, content := range allContent {
		contentMap[content.Slug] = content
	}

	// Function to recursively collect dependencies
	var collectDependencies func(content project.ContentData, visited map[string]bool) []project.ContentData
	collectDependencies = func(content project.ContentData, visited map[string]bool) []project.ContentData {
		if visited[content.Slug] {
			return nil // Avoid circular dependencies
		}
		visited[content.Slug] = true

		var deps []project.ContentData
		deps = append(deps, content)

		// Add required dependencies
		for _, dep := range content.Dependencies {
			if dep.DependencyType == project.Required {
				if depContent, exists := contentMap[dep.Slug]; exists {
					deps = append(deps, collectDependencies(depContent, visited)...)
				}
			}
		}

		return deps
	}

	// Second pass: filter content and include dependencies
	included := make(map[string]bool)

	for _, content := range allContent {
		// Skip if already processed
		if included[content.Slug] {
			continue
		}

		// Check side requirements for the main mod
		includeContent := true
		if serverOnly && content.Side.Server != 2 { // 2 = required
			includeContent = false
		}
		if clientOnly && content.Side.Client != 2 { // 2 = required
			includeContent = false
		}

		// Filter by source
		if source != "" {
			var contentSource string
			switch content.Source {
			case 0:
				contentSource = "modrinth"
			case 1:
				contentSource = "curseforge"
			default:
				contentSource = "unknown"
			}
			if contentSource != source {
				includeContent = false
			}
		}

		if includeContent {
			// Collect this mod and all its required dependencies
			visited := make(map[string]bool)
			dependencies := collectDependencies(content, visited)

			for _, dep := range dependencies {
				if !included[dep.Slug] {
					// For dependencies, we're more lenient with side requirements
					// If the parent mod is included, we include required dependencies regardless of their side
					// unless they're explicitly incompatible
					depInclude := true

					// Still respect source filtering for dependencies
					if source != "" {
						var depSource string
						switch dep.Source {
						case 0:
							depSource = "modrinth"
						case 1:
							depSource = "curseforge"
						default:
							depSource = "unknown"
						}
						if depSource != source {
							depInclude = false
						}
					}

					// Don't include client-only dependencies for server-only filtering
					if serverOnly && dep.Side.Client == 2 && dep.Side.Server == 0 {
						depInclude = false
					}
					// Don't include server-only dependencies for client-only filtering
					if clientOnly && dep.Side.Server == 2 && dep.Side.Client == 0 {
						depInclude = false
					}

					if depInclude {
						filtered = append(filtered, dep)
						included[dep.Slug] = true
					}
				}
			}
		}
	}

	return filtered
}

// linkUpdateCmd represents the link update command
var linkUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update all linked Minecraft instances with current modpack content",
	Long:  `downloads missing mods and syncs overrides to all linked Minecraft instance folders`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		serverOnly, _ := cmd.Flags().GetBool("server-only")
		clientOnly, _ := cmd.Flags().GetBool("client-only")
		source, _ := cmd.Flags().GetString("source")

		// Validate conflicting flags
		if serverOnly && clientOnly {
			fmt.Print(util.FormatError("cannot use --server-only and --client-only together\n"))
			return
		}

		if source != "" && source != "modrinth" && source != "curseforge" {
			fmt.Printf(util.FormatError("invalid source: %s (must be 'modrinth' or 'curseforge')\n"), source)
			return
		}
		// get current working directory and parse project
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf(util.FormatError("error getting current working directory: %s"), err)
			return
		}

		packData, err := project.ParseProject(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error parsing project: %s"), err)
			return
		}

		// Load linked folders
		linked, err := loadLinkedFolders(cwd)
		if err != nil {
			fmt.Printf(util.FormatError("error loading linked folders: %s"), err)
			return
		}

		if len(linked.Links) == 0 {
			fmt.Println("No linked folders found. Use 'minepack link add [folder]' to add links first.")
			return
		}

		// Load link state for tracking removed files and overrides changes
		linkState, err := LoadLinkState(cwd)
		if err != nil {
			fmt.Printf(util.FormatWarning("warning: failed to load link state: %s\n"), err)
			linkState = &LinkState{
				RemovedFiles:   []string{},
				OverridesFiles: make(map[string]string),
				Version:        "1.0",
			}
		}

		// Get all content
		allContent, err := packData.GetAllContent()
		if err != nil {
			fmt.Printf(util.FormatError("error getting all content: %s"), err)
			return
		}

		// Apply filters
		allContent = filterContent(allContent, serverOnly, clientOnly, source)

		if len(allContent) == 0 {
			fmt.Println("no mods found matching the specified filters.")
			return
		}

		// Calculate what files are missing from each link
		var allMissingFiles []project.ContentData
		missingFilesByLink := make(map[string][]project.ContentData)

		for _, linkPath := range linked.Links {
			missingFiles, err := getMissingFiles(linkPath, allContent)
			if err != nil {
				fmt.Printf(util.FormatWarning("error checking files in %s: %s\n"), linkPath, err)
				continue
			}
			missingFilesByLink[linkPath] = missingFiles

			// Add to total missing files (avoiding duplicates)
			for _, missing := range missingFiles {
				found := false
				for _, existing := range allMissingFiles {
					if existing.Slug == missing.Slug {
						found = true
						break
					}
				}
				if !found {
					allMissingFiles = append(allMissingFiles, missing)
				}
			}
		}

		// Calculate download size and total disk usage
		var downloadSize int64
		for _, content := range allMissingFiles {
			downloadSize += content.File.Filesize
		}

		totalDiskUsage := downloadSize * int64(len(linked.Links)+1) // +1 for cache

		// Show summary and ask for confirmation
		var finalString string

		finalString += "update summary:\n"
		finalString += fmt.Sprintf("- linked instances: %d\n", len(linked.Links))
		finalString += fmt.Sprintf("- total mods in modpack: %d\n", len(allContent))
		finalString += fmt.Sprintf("- unique missing files: %d\n", len(allMissingFiles))
		finalString += fmt.Sprintf("- download size: %s\n", formatFileSize(downloadSize))
		finalString += fmt.Sprintf("- total disk usage after sync: %s", formatFileSize(totalDiskUsage))
		fmt.Print(updateSummaryStyle.Render(finalString))
		fmt.Println()

		if len(allMissingFiles) == 0 {
			fmt.Println("all linked instances are up to date!")
		} else {
			// Ask for confirmation
			var proceed bool
			huh.NewConfirm().
				Title("proceed with download and sync?").
				Description("this will download missing files and sync overrides to all linked instances").
				Affirmative("yup!").
				Negative("hell nah").
				Value(&proceed).
				Run()

			if !proceed {
				fmt.Println(util.FormatError("operation cancelled"))
				return
			}

			// Create cache directory
			cacheDir := filepath.Join(cwd, ".mpcache")
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				fmt.Printf(util.FormatError("error creating cache directory: %s"), err)
				return
			}

			// Download missing files to cache
			if len(allMissingFiles) > 0 {
				// Set up parallel download with progress bar
				jobs := make(chan project.ContentData, len(allMissingFiles))
				results := make(chan downloadMsg, len(allMissingFiles))

				// Start 3 worker goroutines
				const numWorkers = 3
				var wg sync.WaitGroup
				for i := 0; i < numWorkers; i++ {
					wg.Add(1)
					go func(workerID int) {
						defer wg.Done()
						downloadWorker(workerID, jobs, results, cacheDir)
					}(i)
				}

				// Send jobs
				for _, content := range allMissingFiles {
					jobs <- content
				}
				close(jobs)

				// Set up progress bar
				prog := newDownloadProgress(len(allMissingFiles))
				p := tea.NewProgram(prog)

				// Monitor results and update progress
				go func() {
					defer func() {
						wg.Wait()
						p.Send(downloadCompleteMsg{})
					}()

					completed := 0
					for completed < len(allMissingFiles) {
						result := <-results
						p.Send(result)
						// Only count completed files, not progress updates
						if result.complete {
							completed++
						}
					}
				}()

				// Run the progress bar
				if _, err := p.Run(); err != nil {
					fmt.Printf(util.FormatError("progress bar error: %s\n"), err)
					return
				}
			}

			// First, delete any tracked removed files from linked instances
			if len(linkState.RemovedFiles) > 0 {
				fmt.Printf("\nremoving %d deleted files from linked instances...\n", len(linkState.RemovedFiles))
				for _, linkPath := range linked.Links {
					removedCount := 0
					for _, removedFile := range linkState.RemovedFiles {
						filePath := filepath.Join(linkPath, removedFile)
						if _, err := os.Stat(filePath); err == nil {
							if err := os.Remove(filePath); err != nil {
								fmt.Printf(util.FormatWarning("failed to remove %s from %s: %s\n"), removedFile, linkPath, err)
							} else {
								removedCount++
							}
						}
					}
					if removedCount > 0 {
						fmt.Printf(util.FormatSuccess("%s (%d files removed)\n"), linkPath, removedCount)
					}
				}
				// Clear the removed files list
				linkState.ClearRemovedFiles()
			}

			// Copy files from cache to each linked instance
			fmt.Println("\nsyncing files to linked instances...")

			for _, linkPath := range linked.Links {
				missingFiles := missingFilesByLink[linkPath]
				if len(missingFiles) == 0 {
					fmt.Printf(util.FormatSuccess("%s (already up to date)\n"), linkPath)
					continue
				}

				// Copy missing files
				successCount := 0
				for _, content := range missingFiles {
					cachePath := filepath.Join(cacheDir, content.File.Filepath)
					destPath := filepath.Join(linkPath, content.File.Filepath)

					// Ensure subdirectory exists
					splitPath := filepath.Dir(destPath)
					if err := os.MkdirAll(splitPath, 0755); err != nil {
						fmt.Printf(util.FormatError("failed to create directory %s: %s\n"), splitPath, err)
						continue
					}

					if err := copyFile(cachePath, destPath); err != nil {
						fmt.Printf(util.FormatError("failed to copy %s to %s: %s\n"), content.Name, linkPath, err)
					} else {
						successCount++
					}
				}

				fmt.Printf(util.FormatSuccess("%s (%d/%d files synced)\n"), linkPath, successCount, len(missingFiles))
			}
		}

		// Sync overrides folder with optimization
		overridesPath := filepath.Join(cwd, "overrides")
		if _, err := os.Stat(overridesPath); err == nil {
			// Get overrides diff to optimize sync
			added, modified, removed, err := linkState.GetOverridesDiff(cwd)
			if err != nil {
				fmt.Printf(util.FormatWarning("warning: failed to calculate overrides diff: %s\n"), err)
				// Fall back to full sync
				fmt.Println("\nsyncing overrides (full sync)...")
				for _, linkPath := range linked.Links {
					if err := copyDir(overridesPath, linkPath); err != nil {
						fmt.Printf(util.FormatError("failed to sync overrides to %s: %s\n"), linkPath, err)
					} else {
						fmt.Printf(util.FormatSuccess("%s (overrides synced)\n"), linkPath)
					}
				}
			} else {
				// Optimized sync based on diff
				totalChanges := len(added) + len(modified) + len(removed)
				if totalChanges == 0 {
					fmt.Println("\noverrides already up to date")
				} else {
					fmt.Printf("\nsyncing overrides (%d added, %d modified, %d removed)...\n", len(added), len(modified), len(removed))

					for _, linkPath := range linked.Links {
						changeCount := 0

						// Remove deleted files
						for _, removedFile := range removed {
							targetPath := filepath.Join(linkPath, removedFile)
							if _, err := os.Stat(targetPath); err == nil {
								if err := os.Remove(targetPath); err != nil {
									fmt.Printf(util.FormatWarning("failed to remove %s from %s: %s\n"), removedFile, linkPath, err)
								} else {
									changeCount++
								}
							}
						}

						// Copy added and modified files
						for _, file := range append(added, modified...) {
							srcPath := filepath.Join(overridesPath, file)
							destPath := filepath.Join(linkPath, file)

							// Ensure subdirectory exists
							destDir := filepath.Dir(destPath)
							if err := os.MkdirAll(destDir, 0755); err != nil {
								fmt.Printf(util.FormatWarning("failed to create directory %s: %s\n"), destDir, err)
								continue
							}

							if err := copyFile(srcPath, destPath); err != nil {
								fmt.Printf(util.FormatWarning("failed to copy %s to %s: %s\n"), file, linkPath, err)
							} else {
								changeCount++
							}
						}

						fmt.Printf(util.FormatSuccess("%s (%d changes applied)\n"), linkPath, changeCount)
					}
				}
			}

			// Update overrides state for next time
			if err := linkState.ScanOverridesFolder(cwd); err != nil {
				fmt.Printf(util.FormatWarning("warning: failed to update overrides state: %s\n"), err)
			}
		} else {
			fmt.Println("\nno overrides folder found, skipping override sync")
		}

		// Save updated link state
		if err := SaveLinkState(cwd, linkState); err != nil {
			fmt.Printf(util.FormatWarning("warning: failed to save link state: %s\n"), err)
		}

		fmt.Println()
		fmt.Print(util.FormatSuccess("sync completed successfully!"))
	},
}

func init() {
	linkCmd.AddCommand(linkUpdateCmd)

	// Add filtering flags
	linkUpdateCmd.Flags().Bool("server-only", false, "only download server-side mods")
	linkUpdateCmd.Flags().Bool("client-only", false, "only download client-side mods")
	linkUpdateCmd.Flags().String("source", "", "only download from specific source (modrinth|curseforge)")
}
