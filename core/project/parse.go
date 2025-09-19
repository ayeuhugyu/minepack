package project

import (
	"errors"
	"fmt"
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

// ParseSumFormat parses a YAML sum file
func ParseSumFormat(sumPath string) (*[]SummaryObject, error) {
	if sumPath == "" {
		return nil, errors.New("summary path is empty")
	}

	if _, err := os.Stat(sumPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sum file does not exist: %s", sumPath)
	}

	sumFile, err := os.Open(sumPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sum file: %w", err)
	}
	defer sumFile.Close()

	// check if file is empty
	fileInfo, err := sumFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		// return empty slice for empty files
		return &[]SummaryObject{}, nil
	}

	var sums []SummaryObject
	decoder := yaml.NewDecoder(sumFile)
	err = decoder.Decode(&sums)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML sum file: %w", err)
	}

	return &sums, nil
}

func ParseSum(projPath string) (*[]SummaryObject, error) {
	if projPath == "" {
		return nil, errors.New("project path is empty")
	}
	var fullPath = projPath + string(filepath.Separator) + "content.mp.sum.yaml"
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("directory is not a minepack project (missing content.mp.sum.yaml)")
	}

	return ParseSumFormat(fullPath)
}

func ParseIncompat(projPath string) (*[]SummaryObject, error) {
	if projPath == "" {
		return nil, errors.New("project path is empty")
	}
	var fullPath = projPath + string(filepath.Separator) + "incompat.mp.sum.yaml"
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, errors.New("directory is not a minepack project (missing incompat.mp.sum.yaml)")
	}

	return ParseSumFormat(fullPath)
}
