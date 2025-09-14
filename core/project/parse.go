package project

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ParseProject(projPath string) (*Project, error) {
	if projPath == "" {
		return nil, errors.New("project path is empty")
	}

	var fullPath = projPath + string(filepath.Separator) + "project.mp.yaml"

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("directory is not a minepack project (missing project.mp.yaml)")
	}

	projFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer projFile.Close()

	var proj Project
	decoder := yaml.NewDecoder(projFile)
	err = decoder.Decode(&proj)
	if err != nil {
		return nil, err
	}

	return &proj, nil
}
