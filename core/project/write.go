package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func WriteProject(proj *Project) error {
	if proj == nil {
		return errors.New("project is nil")
	}

	projDir := strings.ReplaceAll(proj.Root, " ", "-")
	if projDir == "" {
		return errors.New("project root is empty")
	}

	err := os.MkdirAll(projDir, os.ModePerm)
	if err != nil {
		return err
	}

	configPath := filepath.Join(projDir, "project.mp.yaml")

	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := yaml.NewEncoder(configFile)
	encoder.SetIndent(2)
	encoder.Encode(proj)

	// if there is not a content folder, create it

	contentDir := filepath.Join(projDir, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		err = os.MkdirAll(contentDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// if there is not any of the following folders, create them:
	// overrides
	// overrides/mods
	// overrides/config
	// overrides/resourcepacks
	// overrides/shaderpacks

	overridesDir := filepath.Join(projDir, "overrides")
	if _, err := os.Stat(overridesDir); os.IsNotExist(err) {
		err = os.MkdirAll(overridesDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	subDirs := []string{"mods", "config", "resourcepacks", "shaderpacks"}
	for _, subDir := range subDirs {
		fullPath := filepath.Join(overridesDir, subDir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			err = os.MkdirAll(fullPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	// if there is not a content.mp.sum.yaml file, create it
	sumPath := filepath.Join(projDir, "content.mp.sum.yaml")
	if _, err := os.Stat(sumPath); os.IsNotExist(err) {
		sumFile, err := os.Create(sumPath)
		if err != nil {
			return err
		}
		defer sumFile.Close()
	}

	// if there is not an incompat.mp.sum.yaml file, create it
	incompatSumPath := filepath.Join(projDir, "incompat.mp.sum.yaml")
	if _, err := os.Stat(incompatSumPath); os.IsNotExist(err) {
		incompatSumFile, err := os.Create(incompatSumPath)
		if err != nil {
			return err
		}
		defer incompatSumFile.Close()
	}

	// Initialize git repository and version history
	_, err = InitGitRepo(projDir)
	if err != nil {
		return err
	}

	// Initialize version history if it doesn't exist
	versionHistoryPath := filepath.Join(projDir, "versions.mp.yaml")
	if _, err := os.Stat(versionHistoryPath); os.IsNotExist(err) {
		history := &VersionHistory{
			Format:  VersionFormatSemVer,
			Current: "0.1.0",
			Entries: []VersionEntry{},
		}
		err = WriteVersionHistory(projDir, history)
		if err != nil {
			return err
		}
	}

	// Create initial commit
	_, err = CreateVersionCommit(projDir, "Initial minepack project")
	if err != nil {
		return err
	}

	return nil
}

func WriteSumFormat(sums []SummaryObject, sumPath string) error {
	if sumPath == "" {
		return errors.New("summary path is empty")
	}

	sumFile, err := os.Create(sumPath)
	if err != nil {
		return fmt.Errorf("failed to create sum file: %w", err)
	}
	defer sumFile.Close()

	encoder := yaml.NewEncoder(sumFile)
	defer encoder.Close()

	err = encoder.Encode(sums)
	if err != nil {
		return fmt.Errorf("failed to write YAML sum file: %w", err)
	}

	return nil
}

func WriteSum(sums []SummaryObject, projPath string) error {
	if projPath == "" {
		return errors.New("project path is empty")
	}

	var fullPath = projPath + string(filepath.Separator) + "content.mp.sum.yaml"
	return WriteSumFormat(sums, fullPath)
}

func WriteIncompatSum(incompatSums []SummaryObject, projPath string) error {
	if projPath == "" {
		return errors.New("project path is empty")
	}

	var fullPath = projPath + string(filepath.Separator) + "incompat.mp.sum.yaml"
	return WriteSumFormat(incompatSums, fullPath)
}
