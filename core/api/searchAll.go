package api

import (
	"minepack/core/api/curseforge"
	"minepack/core/api/modrinth"
	"minepack/core/project"

	modrinthApi "codeberg.org/jmansfield/go-modrinth/modrinth"

	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

func extractSlugFromURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL")
	}
	// Modrinth: https://modrinth.com/mod/<slug>
	if strings.Contains(url, "modrinth.com/mod/") {
		parts := strings.Split(url, "/")
		for i, part := range parts {
			if part == "mod" && i > 0 && parts[i-1] == "modrinth.com" {
				if i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
			if part == "mod" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		// fallback: last part
		return parts[len(parts)-1], nil
	}

	// CurseForge: https://www.curseforge.com/minecraft/mc-mods/<slug>
	if strings.Contains(url, "curseforge.com/minecraft/mc-mods/") {
		parts := strings.Split(url, "/")
		for i, part := range parts {
			if part == "mc-mods" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		// fallback: last part
		return parts[len(parts)-1], nil
	}

	// Not a supported URL
	return "", nil
}

func findBySlugMatch(query string, packData project.Project) (*project.ContentData, error) {
	var defaultSource = packData.DefaultSource
	if defaultSource == "" {
		defaultSource = "modrinth"
	}

	var result project.ContentData
	var foundResult bool = false
	if defaultSource == "modrinth" {
		mrresult, err := modrinth.GetProject(query)
		if err != nil || mrresult == nil {
			return nil, nil
		}
		mrversions, err := modrinth.GetProjectVersions(*mrresult.ID, packData)
		if err != nil || len(mrversions) == 0 {
			return nil, nil
		}
		result = modrinth.ConvertProjectToContentData(mrresult, mrversions[0])
		foundResult = true
	}
	if defaultSource == "curseforge" {
		cfresult, err := curseforge.GetProject(query)
		if err != nil || cfresult == nil {
			return nil, nil
		}
		var stringid string = fmt.Sprintf("%d", cfresult.ID)
		cfversions, err := curseforge.GetProjectVersions(stringid, packData)
		if err != nil || len(cfversions) == 0 {
			return nil, nil
		}
		result = curseforge.ConvertModToContentData(cfresult, &cfversions[0])
		foundResult = true
	}
	if foundResult {
		return &result, nil
	}
	return nil, nil
}

func findBySearch(query string, packData project.Project) (*project.ContentData, error) {
	var defaultSource = packData.DefaultSource
	if defaultSource == "" {
		defaultSource = "modrinth"
	}

	var result *project.ContentData
	var resultFound bool = false

	if defaultSource == "modrinth" {
		mrresults, err := modrinth.SearchProjects(query, packData, false)
		if err != nil {
			return nil, err
		}
		if len(mrresults) == 1 {
			mrversions, err := modrinth.GetProjectVersions(*mrresults[0].Slug, packData)
			if err != nil || len(mrversions) == 0 {
				return nil, err
			}
			mrproject, err := modrinth.GetProject(*mrresults[0].Slug)
			if err != nil || mrproject == nil {
				return nil, err
			}
			resultData := modrinth.ConvertProjectToContentData(mrproject, mrversions[0])
			result = &resultData
			resultFound = true
		}
		if len(mrresults) > 1 {
			// if multiple results, use huh to ask the user to pick one
			var opts []huh.Option[*modrinthApi.SearchResult]

			for _, r := range mrresults {
				opts = append(opts, huh.NewOption(
					fmt.Sprintf("%s \033[90m(%s)\033[0m", *r.Title, *r.Slug), r),
				)
			}

			var chosen *modrinthApi.SearchResult

			searchResultsForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[*modrinthApi.SearchResult]().
						Title("multiple modrinth search results found").
						Description("please select the best matching mod (or press ctrl+c to cancel)").
						Options(opts...).
						Value(&chosen),
				),
			)

			err := searchResultsForm.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt failed %v", err)
			}

			mrversions, err := modrinth.GetProjectVersions(*chosen.Slug, packData)
			if err != nil || len(mrversions) == 0 {
				return nil, err
			}
			mrproject, err := modrinth.GetProject(*chosen.Slug)
			if err != nil || mrproject == nil {
				return nil, err
			}
			resultData := modrinth.ConvertProjectToContentData(mrproject, mrversions[0])
			result = &resultData
			resultFound = true
		}
	}
	if defaultSource == "curseforge" {
		cfresults, err := curseforge.SearchProjects(query, packData, false)
		if err != nil {
			return nil, err
		}
		if len(cfresults) == 1 {
			var stringid string = fmt.Sprintf("%d", cfresults[0].ID)
			cfversions, err := curseforge.GetProjectVersions(stringid, packData)
			if err != nil || len(cfversions) == 0 {
				return nil, err
			}
			cfproject, err := curseforge.GetProject(stringid)
			if err != nil || cfproject == nil {
				return nil, err
			}
			resultData := curseforge.ConvertModToContentData(cfproject, &cfversions[0])
			result = &resultData
			resultFound = true
		}
		if len(cfresults) > 1 {
			// if multiple results, use huh to ask the user to pick one
			var opts []huh.Option[*curseforge.Mod]

			for _, r := range cfresults {
				opts = append(opts, huh.NewOption(r.Name, &r))
			}

			var chosen *curseforge.Mod

			searchResultsForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[*curseforge.Mod]().
						Title("multiple curseforge search results found").
						Description("please select the best matching mod").
						Options(opts...).
						Value(&chosen),
				),
			)

			err := searchResultsForm.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt failed %v", err)
			}

			var stringid string = fmt.Sprintf("%d", chosen.ID)
			cfversions, err := curseforge.GetProjectVersions(stringid, packData)
			if err != nil || len(cfversions) == 0 {
				return nil, err
			}
			cfproject, err := curseforge.GetProject(stringid)
			if err != nil || cfproject == nil {
				return nil, err
			}
			resultData := curseforge.ConvertModToContentData(cfproject, &cfversions[0])
			result = &resultData
			resultFound = true
		}
	}
	if resultFound {
		return result, nil
	}
	return nil, nil
}

func SearchAll(query string, packData project.Project) (*project.ContentData, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}
	// first, see if the input is a URL and if so, extract the slug from it
	slug, err := extractSlugFromURL(query)
	if err != nil {
		return nil, err
	}
	if slug != "" {
		query = slug
	}

	// second, try directly matching the slug with the project's default source
	result, err := findBySlugMatch(query, packData)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	// third, search default source for the query
	result, err = findBySearch(query, packData)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	// no results found
	return nil, nil
}
