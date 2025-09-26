package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// LinkState represents the state of linked instances
type LinkState struct {
	LastUpdate      time.Time         `yaml:"last_update"`
	RemovedFiles    []string          `yaml:"removed_files"`    // file paths of removed mods/content
	OverridesFiles  map[string]string `yaml:"overrides_files"`  // path -> hash of files in overrides
	Version         string            `yaml:"version"`
}

// getLinkStateFile returns the path to linkstate.mp.yaml
func getLinkStateFile(projectRoot string) string {
	return filepath.Join(projectRoot, "linkstate.mp.yaml")
}

// LoadLinkState loads the link state from linkstate.mp.yaml
func LoadLinkState(projectRoot string) (*LinkState, error) {
	stateFile := getLinkStateFile(projectRoot)

	// If file doesn't exist, return empty state
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return &LinkState{
			RemovedFiles:   []string{},
			OverridesFiles: make(map[string]string),
			Version:        "1.0",
		}, nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read linkstate.mp.yaml: %w", err)
	}

	var state LinkState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse linkstate.mp.yaml: %w", err)
	}

	// Initialize maps if nil
	if state.RemovedFiles == nil {
		state.RemovedFiles = []string{}
	}
	if state.OverridesFiles == nil {
		state.OverridesFiles = make(map[string]string)
	}

	return &state, nil
}

// SaveLinkState saves the link state to linkstate.mp.yaml
func SaveLinkState(projectRoot string, state *LinkState) error {
	stateFile := getLinkStateFile(projectRoot)

	state.LastUpdate = time.Now()

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal link state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write linkstate.mp.yaml: %w", err)
	}

	return nil
}

// AddRemovedFile adds a file path to the list of removed files
func (ls *LinkState) AddRemovedFile(filepath string) {
	// Check if already exists
	for _, existing := range ls.RemovedFiles {
		if existing == filepath {
			return
		}
	}
	ls.RemovedFiles = append(ls.RemovedFiles, filepath)
}

// ClearRemovedFiles clears the list of removed files
func (ls *LinkState) ClearRemovedFiles() {
	ls.RemovedFiles = []string{}
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ScanOverridesFolder scans the overrides folder and updates the file hash map
func (ls *LinkState) ScanOverridesFolder(projectRoot string) error {
	overridesPath := filepath.Join(projectRoot, "overrides")
	
	// Check if overrides folder exists
	if _, err := os.Stat(overridesPath); os.IsNotExist(err) {
		// No overrides folder, clear the map
		ls.OverridesFiles = make(map[string]string)
		return nil
	}

	newOverridesFiles := make(map[string]string)

	err := filepath.Walk(overridesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from overrides folder
		relPath, err := filepath.Rel(overridesPath, path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash, err := calculateFileHash(path)
		if err != nil {
			return err
		}

		newOverridesFiles[relPath] = hash
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan overrides folder: %w", err)
	}

	ls.OverridesFiles = newOverridesFiles
	return nil
}

// GetOverridesDiff returns lists of added, modified, and removed files in overrides
func (ls *LinkState) GetOverridesDiff(projectRoot string) (added []string, modified []string, removed []string, err error) {
	// Get current state of overrides folder
	currentState := &LinkState{OverridesFiles: make(map[string]string)}
	if err := currentState.ScanOverridesFolder(projectRoot); err != nil {
		return nil, nil, nil, err
	}

	// Compare with previous state
	for relPath, currentHash := range currentState.OverridesFiles {
		if previousHash, exists := ls.OverridesFiles[relPath]; exists {
			if previousHash != currentHash {
				modified = append(modified, relPath)
			}
		} else {
			added = append(added, relPath)
		}
	}

	// Find removed files
	for relPath := range ls.OverridesFiles {
		if _, exists := currentState.OverridesFiles[relPath]; !exists {
			removed = append(removed, relPath)
		}
	}

	return added, modified, removed, nil
}