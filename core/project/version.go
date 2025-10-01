package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"
)

// VersionFormat represents the version format type
type VersionFormat string

const (
	VersionFormatSemVer    VersionFormat = "semver"
	VersionFormatBreakVer  VersionFormat = "breakver"
	VersionFormatIncrement VersionFormat = "increment"
	VersionFormatCustom    VersionFormat = "custom"
)

// VersionEntry represents a single version entry in the history
type VersionEntry struct {
	Version   string    `yaml:"version"`
	CommitSHA string    `yaml:"commit_sha"`
	Message   string    `yaml:"message"`
	Timestamp time.Time `yaml:"timestamp"`
}

// VersionHistory stores version history and format
type VersionHistory struct {
	Format  VersionFormat  `yaml:"format"`
	Current string         `yaml:"current"`
	Entries []VersionEntry `yaml:"entries"`
}

// ParseVersionHistory reads the versions.mp.yaml file
func ParseVersionHistory(projPath string) (*VersionHistory, error) {
	versionPath := filepath.Join(projPath, "versions.mp.yaml")
	
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		// Return default version history if file doesn't exist
		return &VersionHistory{
			Format:  VersionFormatSemVer,
			Current: "0.1.0",
			Entries: []VersionEntry{},
		}, nil
	}

	versionFile, err := os.Open(versionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open versions file: %w", err)
	}
	defer versionFile.Close()

	var history VersionHistory
	decoder := yaml.NewDecoder(versionFile)
	if err := decoder.Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to parse versions file: %w", err)
	}

	return &history, nil
}

// WriteVersionHistory writes the version history to versions.mp.yaml
func WriteVersionHistory(projPath string, history *VersionHistory) error {
	versionPath := filepath.Join(projPath, "versions.mp.yaml")
	
	versionFile, err := os.Create(versionPath)
	if err != nil {
		return fmt.Errorf("failed to create versions file: %w", err)
	}
	defer versionFile.Close()

	encoder := yaml.NewEncoder(versionFile)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(history); err != nil {
		return fmt.Errorf("failed to write versions file: %w", err)
	}

	return nil
}

// InitGitRepo initializes a git repository in the project directory
func InitGitRepo(projPath string) (*git.Repository, error) {
	// Check if .git already exists
	gitPath := filepath.Join(projPath, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		// Repository already exists, open it
		return git.PlainOpen(projPath)
	}

	// Initialize new repository
	repo, err := git.PlainInit(projPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	return repo, nil
}

// CreateVersionCommit creates a git commit for a version change
func CreateVersionCommit(projPath string, message string) (string, error) {
	repo, err := git.PlainOpen(projPath)
	// if the error is that the repository does not exist, initialize it
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = InitGitRepo(projPath)
	}
	// if there is still an error, return it
	if err != nil {
		return "", fmt.Errorf("failed to open git repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all files
	if err := worktree.AddGlob("."); err != nil {
		return "", fmt.Errorf("failed to add files: %w", err)
	}

	// Create commit
	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Minepack",
			Email: "minepack@local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %w", err)
	}

	return commit.String(), nil
}

// AutoCommit creates an automatic commit for content changes
func AutoCommit(projPath string, message string) error {
	_, err := CreateVersionCommit(projPath, message)
	return err
}

// ParseSemVer parses a semantic version string
func ParseSemVer(version string) (major, minor, patch int, err error) {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	
	if len(parts) != 3 {
		return 0, 0, 0, errors.New("invalid semver format, expected Major.Minor.Patch")
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// FormatSemVer formats major, minor, patch as a semantic version string
func FormatSemVer(major, minor, patch int) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// UpdateSemVerMajor updates the major version
func UpdateSemVerMajor(version string, operation string, value int) (string, error) {
	major, minor, patch, err := ParseSemVer(version)
	if err != nil {
		return "", err
	}

	switch operation {
	case "add":
		major += value
	case "subtract":
		major -= value
		if major < 0 {
			major = 0
		}
	case "set":
		major = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	// Reset minor and patch when incrementing major
	if operation == "add" && value > 0 {
		minor = 0
		patch = 0
	}

	return FormatSemVer(major, minor, patch), nil
}

// UpdateSemVerMinor updates the minor version
func UpdateSemVerMinor(version string, operation string, value int) (string, error) {
	major, minor, patch, err := ParseSemVer(version)
	if err != nil {
		return "", err
	}

	switch operation {
	case "add":
		minor += value
	case "subtract":
		minor -= value
		if minor < 0 {
			minor = 0
		}
	case "set":
		minor = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	// Reset patch when incrementing minor
	if operation == "add" && value > 0 {
		patch = 0
	}

	return FormatSemVer(major, minor, patch), nil
}

// UpdateSemVerPatch updates the patch version
func UpdateSemVerPatch(version string, operation string, value int) (string, error) {
	major, minor, patch, err := ParseSemVer(version)
	if err != nil {
		return "", err
	}

	switch operation {
	case "add":
		patch += value
	case "subtract":
		patch -= value
		if patch < 0 {
			patch = 0
		}
	case "set":
		patch = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	return FormatSemVer(major, minor, patch), nil
}

// ParseBreakVer parses a breaking version string (MAJOR.MINOR)
func ParseBreakVer(version string) (major, minor int, err error) {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid breakver format, expected Major.Minor")
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	return major, minor, nil
}

// FormatBreakVer formats major, minor as a breaking version string
func FormatBreakVer(major, minor int) string {
	return fmt.Sprintf("%d.%d", major, minor)
}

// UpdateBreakVerMajor updates the major version (breakver format)
func UpdateBreakVerMajor(version string, operation string, value int) (string, error) {
	major, minor, err := ParseBreakVer(version)
	if err != nil {
		return "", err
	}

	switch operation {
	case "add":
		major += value
	case "subtract":
		major -= value
		if major < 0 {
			major = 0
		}
	case "set":
		major = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	// Reset minor when incrementing major
	if operation == "add" && value > 0 {
		minor = 0
	}

	return FormatBreakVer(major, minor), nil
}

// UpdateBreakVerMinor updates the minor version (breakver format)
func UpdateBreakVerMinor(version string, operation string, value int) (string, error) {
	major, minor, err := ParseBreakVer(version)
	if err != nil {
		return "", err
	}

	switch operation {
	case "add":
		minor += value
	case "subtract":
		minor -= value
		if minor < 0 {
			minor = 0
		}
	case "set":
		minor = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	return FormatBreakVer(major, minor), nil
}

// UpdateIncrementVersion updates an increment-based version
func UpdateIncrementVersion(version string, operation string, value int) (string, error) {
	current, err := strconv.Atoi(version)
	if err != nil {
		return "", fmt.Errorf("invalid increment version: %w", err)
	}

	switch operation {
	case "add":
		current += value
	case "subtract":
		current -= value
		if current < 0 {
			current = 0
		}
	case "set":
		current = value
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	return strconv.Itoa(current), nil
}

// SetVersion sets the project version and creates a commit
func SetVersion(projPath string, newVersion string, message string) error {
	history, err := ParseVersionHistory(projPath)
	if err != nil {
		return err
	}

	// Update history first to ensure there's a change to commit
	entry := VersionEntry{
		Version:   newVersion,
		CommitSHA: "",  // Will be filled after commit
		Message:   message,
		Timestamp: time.Now(),
	}
	history.Entries = append(history.Entries, entry)
	history.Current = newVersion

	// Write updated history
	err = WriteVersionHistory(projPath, history)
	if err != nil {
		return err
	}

	// Create git commit
	commitSHA, err := CreateVersionCommit(projPath, message)
	if err != nil {
		return err
	}

	// Update the entry with the commit SHA
	history.Entries[len(history.Entries)-1].CommitSHA = commitSHA

	// Write updated history again with commit SHA
	return WriteVersionHistory(projPath, history)
}

// RevertToVersion reverts to a specific version by checking out its commit
func RevertToVersion(projPath string, targetVersion string) error {
	history, err := ParseVersionHistory(projPath)
	if err != nil {
		return err
	}

	// Find the target version in history
	var targetEntry *VersionEntry
	for i := range history.Entries {
		if history.Entries[i].Version == targetVersion {
			targetEntry = &history.Entries[i]
			break
		}
	}

	if targetEntry == nil {
		return fmt.Errorf("version %s not found in history", targetVersion)
	}

	// Open repository
	repo, err := git.PlainOpen(projPath)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Parse commit hash
	hash, err := repo.ResolveRevision(plumbing.Revision(targetEntry.CommitSHA))
	if err != nil {
		return fmt.Errorf("failed to resolve commit: %w", err)
	}

	// Checkout the commit
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout commit: %w", err)
	}

	// Update current version
	history.Current = targetVersion
	return WriteVersionHistory(projPath, history)
}

// SetVersionFormat changes the version format
func SetVersionFormat(projPath string, format VersionFormat) error {
	history, err := ParseVersionHistory(projPath)
	if err != nil {
		return err
	}

	history.Format = format
	
	// Set appropriate default version based on format
	if history.Current == "" {
		switch format {
		case VersionFormatSemVer:
			history.Current = "0.1.0"
		case VersionFormatBreakVer:
			history.Current = "0.1"
		case VersionFormatIncrement:
			history.Current = "1"
		case VersionFormatCustom:
			history.Current = "1.0"
		}
	}

	// change the current version to match the default format if it doesn't match
	switch format {
	case VersionFormatSemVer:
		if _, _, _, err := ParseSemVer(history.Current); err != nil {
			history.Current = "0.1.0"
		}
	case VersionFormatBreakVer:
		if _, _, err := ParseBreakVer(history.Current); err != nil {
			history.Current = "0.1"
		}
	case VersionFormatIncrement:
		if _, err := strconv.Atoi(history.Current); err != nil {
			history.Current = "1"
		}
	case VersionFormatCustom:
		// Custom format, do not enforce any specific version pattern
		if history.Current == "" {
			history.Current = "1.0"
		}
	}

	return WriteVersionHistory(projPath, history)
}
