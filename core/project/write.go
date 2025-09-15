package project

import (
	"errors"
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

	// if there is not a mods folder, create it

	modsDir := filepath.Join(projDir, "mods")
	if _, err := os.Stat(modsDir); os.IsNotExist(err) {
		err = os.MkdirAll(modsDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
