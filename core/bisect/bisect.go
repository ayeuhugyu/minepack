package bisect

import (
	"fmt"
	"minepack/core/project"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// BisectState represents the current state of a bisection
type BisectState struct {
	LinkedInstance string              `yaml:"linked_instance"`
	AllMods        []string            `yaml:"all_mods"` // slugs of all mods being bisected
	History        []BisectStep        `yaml:"history"`
	CurrentStep    int                 `yaml:"current_step"`
	Dependencies   map[string][]string `yaml:"dependencies"` // slug -> list of dependent slugs
	ModFiles       map[string]string   `yaml:"mod_files"`    // slug -> filename
	Created        string              `yaml:"created"`
}

// BisectStep represents one step in the bisection process
type BisectStep struct {
	DisabledMods []string `yaml:"disabled_mods"` // slugs of mods disabled in this step
	EnabledMods  []string `yaml:"enabled_mods"`  // slugs of mods enabled in this step
	TestResult   string   `yaml:"test_result"`   // "good", "bad", or "unknown"
}

// getBisectFile returns the path to bisect.mp.yaml
func getBisectFile(projectRoot string) string {
	return filepath.Join(projectRoot, "bisect.mp.yaml")
}

// LoadBisectState loads the current bisection state from bisect.mp.yaml
func LoadBisectState(projectRoot string) (*BisectState, error) {
	bisectFile := getBisectFile(projectRoot)

	if _, err := os.Stat(bisectFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("no active bisection found")
	}

	data, err := os.ReadFile(bisectFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read bisect.mp.yaml: %w", err)
	}

	var state BisectState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse bisect.mp.yaml: %w", err)
	}

	return &state, nil
}

// SaveBisectState saves the bisection state to bisect.mp.yaml
func SaveBisectState(projectRoot string, state *BisectState) error {
	bisectFile := getBisectFile(projectRoot)

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal bisect state: %w", err)
	}

	if err := os.WriteFile(bisectFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write bisect.mp.yaml: %w", err)
	}

	return nil
}

// DeleteBisectState removes the bisect.mp.yaml file
func DeleteBisectState(projectRoot string) error {
	bisectFile := getBisectFile(projectRoot)
	if _, err := os.Stat(bisectFile); err == nil {
		return os.Remove(bisectFile)
	}
	return nil
}

// buildDependencyMap creates a map of mod -> list of mods that depend on it
func buildDependencyMap(allContent []project.ContentData) map[string][]string {
	dependencies := make(map[string][]string)

	for _, content := range allContent {
		for _, dep := range content.Dependencies {
			// Add this mod as dependent on the dependency
			depSlug := dep.Slug
			if dependencies[depSlug] == nil {
				dependencies[depSlug] = []string{}
			}
			dependencies[depSlug] = append(dependencies[depSlug], content.Slug)
		}
	}

	return dependencies
}

// getModsToDisable returns the list of mods that should be disabled when disabling a specific mod
// This includes the mod itself and all mods that depend on it (recursively)
func getModsToDisable(modSlug string, dependencies map[string][]string, visited map[string]bool) []string {
	if visited[modSlug] {
		return []string{}
	}

	visited[modSlug] = true
	result := []string{modSlug}

	// Add all mods that depend on this mod
	for _, dependent := range dependencies[modSlug] {
		result = append(result, getModsToDisable(dependent, dependencies, visited)...)
	}

	return result
}

// CreateBisectState initializes a new bisection with all available mods
func CreateBisectState(projectRoot, linkedInstance string, allContent []project.ContentData) (*BisectState, error) {
	// Build list of all mod slugs and their filenames
	allMods := make([]string, 0, len(allContent))
	modFiles := make(map[string]string)

	for _, content := range allContent {
		allMods = append(allMods, content.Slug)
		modFiles[content.Slug] = filepath.Base(content.File.Filepath)
	}

	// Build dependency map
	dependencies := buildDependencyMap(allContent)

	state := &BisectState{
		LinkedInstance: linkedInstance,
		AllMods:        allMods,
		History:        []BisectStep{},
		CurrentStep:    -1,
		Dependencies:   dependencies,
		ModFiles:       modFiles,
		Created:        "today", // You might want to use time.Now().Format() here
	}

	return state, nil
}

// GetNextBisectStep calculates which mods should be disabled in the next step
func (bs *BisectState) GetNextBisectStep() ([]string, []string, error) {
	// Get the current set of "candidate" mods (mods that could be causing the issue)
	candidates := bs.GetCurrentCandidates()

	if len(candidates) <= 1 {
		return nil, nil, fmt.Errorf("bisection complete")
	}

	// Use a smarter algorithm that accounts for dependency explosion
	toDisable := bs.selectModsForDisabling(candidates)

	// Expand to include dependent mods
	visited := make(map[string]bool)
	expandedDisabled := []string{}
	for _, modSlug := range toDisable {
		expandedDisabled = append(expandedDisabled, getModsToDisable(modSlug, bs.Dependencies, visited)...)
	}

	// Remove duplicates
	disabledMap := make(map[string]bool)
	for _, mod := range expandedDisabled {
		disabledMap[mod] = true
	}

	finalDisabled := make([]string, 0, len(disabledMap))
	for mod := range disabledMap {
		finalDisabled = append(finalDisabled, mod)
	}

	// Enabled mods are all mods not in the disabled list
	enabled := []string{}
	for _, mod := range bs.AllMods {
		if !disabledMap[mod] {
			enabled = append(enabled, mod)
		}
	}

	return finalDisabled, enabled, nil
}

// selectModsForDisabling chooses which mods to disable, trying to get close to 50% after dependency expansion
func (bs *BisectState) selectModsForDisabling(candidates []string) []string {
	targetDisabled := len(candidates) / 2
	if targetDisabled == 0 {
		targetDisabled = 1
	}

	// Calculate dependency impact for each mod
	type modImpact struct {
		mod   string
		count int
	}

	impacts := []modImpact{}
	for _, mod := range candidates {
		visited := make(map[string]bool)
		dependents := getModsToDisable(mod, bs.Dependencies, visited)
		// Only count dependents that are also in candidates
		candidateSet := make(map[string]bool)
		for _, c := range candidates {
			candidateSet[c] = true
		}

		actualCount := 0
		for _, dep := range dependents {
			if candidateSet[dep] {
				actualCount++
			}
		}

		impacts = append(impacts, modImpact{mod: mod, count: actualCount})
	}

	// Sort by impact (smallest first, so we can build up to target)
	for i := 0; i < len(impacts)-1; i++ {
		for j := i + 1; j < len(impacts); j++ {
			if impacts[i].count > impacts[j].count {
				impacts[i], impacts[j] = impacts[j], impacts[i]
			}
		}
	}

	// Greedily select mods until we get close to target
	selected := []string{}
	totalImpact := 0
	usedMods := make(map[string]bool)

	for _, impact := range impacts {
		// Check if this mod is already covered by a previous selection
		if usedMods[impact.mod] {
			continue
		}

		// If adding this would put us way over target, check if we should stop
		if totalImpact+impact.count > targetDisabled*2 && totalImpact > 0 {
			break
		}

		// Add this mod
		selected = append(selected, impact.mod)

		// Mark all mods that would be disabled by this selection as used
		visited := make(map[string]bool)
		dependents := getModsToDisable(impact.mod, bs.Dependencies, visited)
		candidateSet := make(map[string]bool)
		for _, c := range candidates {
			candidateSet[c] = true
		}

		for _, dep := range dependents {
			if candidateSet[dep] {
				usedMods[dep] = true
			}
		}

		totalImpact += impact.count

		// If we've hit our target, stop
		if totalImpact >= targetDisabled {
			break
		}
	}

	// If we haven't selected anything yet, just select the first mod
	if len(selected) == 0 && len(candidates) > 0 {
		selected = append(selected, candidates[0])
	}

	return selected
}

// GetCurrentCandidates returns the list of mods that could currently be causing the issue
func (bs *BisectState) GetCurrentCandidates() []string {
	if len(bs.History) == 0 {
		return bs.AllMods
	}

	// Start with all mods as candidates
	candidates := make(map[string]bool)
	for _, mod := range bs.AllMods {
		candidates[mod] = true
	}

	// Process history to narrow down candidates
	for _, step := range bs.History {
		if step.TestResult == "good" {
			// If test was good with these mods disabled, the problem is NOT in the disabled mods
			// So remove disabled mods from candidates
			for _, mod := range step.DisabledMods {
				delete(candidates, mod)
			}
		} else if step.TestResult == "bad" {
			// If test was bad with these mods disabled, the problem IS in the disabled mods
			// So remove enabled mods from candidates
			for _, mod := range step.EnabledMods {
				delete(candidates, mod)
			}
		}
	}

	result := make([]string, 0, len(candidates))
	for mod := range candidates {
		result = append(result, mod)
	}

	return result
}

// AddStepResult adds the result of testing the current step
func (bs *BisectState) AddStepResult(result string) error {
	if bs.CurrentStep < 0 || bs.CurrentStep >= len(bs.History) {
		return fmt.Errorf("no current step to add result to")
	}

	bs.History[bs.CurrentStep].TestResult = result
	return nil
}

// GoToPreviousStep moves back to the previous step in the bisection
func (bs *BisectState) GoToPreviousStep() error {
	if bs.CurrentStep <= 0 {
		return fmt.Errorf("already at the first step")
	}

	bs.CurrentStep--
	return nil
}

// AddStep adds a new step to the bisection
func (bs *BisectState) AddStep(disabled, enabled []string) {
	step := BisectStep{
		DisabledMods: disabled,
		EnabledMods:  enabled,
		TestResult:   "unknown",
	}

	bs.History = append(bs.History, step)
	bs.CurrentStep = len(bs.History) - 1
}

// ApplyCurrentStep applies the current step by renaming mod files
func (bs *BisectState) ApplyCurrentStep() error {
	if bs.CurrentStep < 0 || bs.CurrentStep >= len(bs.History) {
		return fmt.Errorf("no current step to apply")
	}

	currentStep := bs.History[bs.CurrentStep]

	// Enable all mods first (remove .disabled extension)
	modsDir := filepath.Join(bs.LinkedInstance, "mods")
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return fmt.Errorf("failed to read mods directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".disabled") {
			oldPath := filepath.Join(modsDir, entry.Name())
			newPath := filepath.Join(modsDir, strings.TrimSuffix(entry.Name(), ".disabled"))
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to enable mod %s: %w", entry.Name(), err)
			}
		}
	}

	// Now disable the mods that should be disabled
	for _, modSlug := range currentStep.DisabledMods {
		filename, exists := bs.ModFiles[modSlug]
		if !exists {
			continue
		}

		modPath := filepath.Join(modsDir, filename)
		disabledPath := modPath + ".disabled"

		if _, err := os.Stat(modPath); err == nil {
			if err := os.Rename(modPath, disabledPath); err != nil {
				return fmt.Errorf("failed to disable mod %s: %w", filename, err)
			}
		}
	}

	return nil
}

// RestoreAllMods enables all mods (removes .disabled extensions)
func (bs *BisectState) RestoreAllMods() error {
	modsDir := filepath.Join(bs.LinkedInstance, "mods")
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return fmt.Errorf("failed to read mods directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".disabled") {
			oldPath := filepath.Join(modsDir, entry.Name())
			newPath := filepath.Join(modsDir, strings.TrimSuffix(entry.Name(), ".disabled"))
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to enable mod %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
