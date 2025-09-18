package curseforge

import (
	"fmt"
	"minepack/core/project"
	"net/url"
	"strconv"
)

// searches curseforge for projects matching the query and optional pack data
func SearchProjects(query string, projectData project.Project, verbose bool) ([]Mod, error) {
	if verbose {
		fmt.Printf("starting CurseForge search for query: %s\n", query)
		if projectData.Versions.Game != "" || projectData.Versions.Loader.Name != "" {
			fmt.Printf("pack data provided: gameVersion=%s, modloader=%s\n", projectData.Versions.Game, projectData.Versions.Loader.Name)
		}
	}

	// build query parameters
	params := url.Values{}
	params.Set("gameId", strconv.Itoa(GameMinecraft))
	if query != "" {
		params.Set("searchFilter", query)
	}
	params.Set("pageSize", "10")
	params.Set("index", "0")

	// add class filter (project types) - only add mods for now to test
	if projectData != (project.Project{}) {
		// start with just mods to test
		params.Add("classId", strconv.Itoa(ClassMods))
		if verbose {
			fmt.Printf("filtering by project type: mods\n")
		}

		// add game version filter
		if projectData.Versions.Game != "" {
			params.Set("gameVersion", projectData.Versions.Game)
			if verbose {
				fmt.Printf("filtering by game version: %s\n", projectData.Versions.Game)
			}
		}

		// add modloader filter
		if projectData.Versions.Loader.Name != "" {
			modLoaderID := getModLoaderID(projectData.Versions.Loader.Name)
			if modLoaderID != ModLoaderAny {
				params.Set("modLoaderType", strconv.Itoa(modLoaderID))
				if verbose {
					fmt.Printf("filtering by modloader: %s (ID: %d)\n", projectData.Versions.Loader.Name, modLoaderID)
				}
			}
		}
	} else {
		// default to mods only for simple search
		params.Add("classId", strconv.Itoa(ClassMods))
	}

	endpoint := "/mods/search?" + params.Encode()
	if verbose {
		fmt.Printf("curseforge API endpoint: %s\n", endpoint)
	}

	var response SearchResponse
	if err := CurseForgeClient.makeRequest(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to search curseforge: %w", err)
	}

	if verbose {
		fmt.Printf("curseforge search completed: %d results found\n", len(response.Data))
	}

	return response.Data, nil
}

// getModLoaderID converts modloader name to CurseForge ID
func getModLoaderID(loaderName string) int {
	switch loaderName {
	case "forge":
		return ModLoaderForge
	case "fabric":
		return ModLoaderFabric
	case "quilt":
		return ModLoaderQuilt
	case "neoforge":
		return ModLoaderNeoForge
	case "liteloader":
		return ModLoaderLiteLoader
	case "cauldron":
		return ModLoaderCauldron
	default:
		return ModLoaderAny
	}
}
