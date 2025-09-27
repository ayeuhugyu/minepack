package curseforge

import (
	"fmt"
	"minepack/core/project"
	"net/url"
	"strconv"
	"strings"
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
	params.Set("pageSize", "50") // Increase page size to get more results
	params.Set("index", "0")
	
	// Add sort by Popularity to get more relevant results first
	params.Set("sortField", "6") // 6 = Popularity  
	params.Set("sortOrder", "desc")

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

	// Post-process results to prioritize exact matches
	results := response.Data
	if len(results) > 1 {
		results = prioritizeExactMatches(results, query, verbose)
	}

	return results, nil
}

// prioritizeExactMatches reorders search results to put exact or close matches first
func prioritizeExactMatches(results []Mod, query string, verbose bool) []Mod {
	if len(results) <= 1 {
		return results
	}
	
	var exactMatches []Mod
	var closeMatches []Mod  
	var otherResults []Mod
	
	queryLower := strings.ToLower(query)
	
	for _, mod := range results {
		modNameLower := strings.ToLower(mod.Name)
		modSlugLower := strings.ToLower(mod.Slug)
		
		// Check for exact name or slug match
		if modNameLower == queryLower || modSlugLower == queryLower {
			exactMatches = append(exactMatches, mod)
			if verbose {
				fmt.Printf("found exact match: %s (slug: %s)\n", mod.Name, mod.Slug)
			}
		} else if strings.Contains(modNameLower, queryLower) || strings.Contains(modSlugLower, queryLower) {
			// Check if query is contained in name or slug
			closeMatches = append(closeMatches, mod)
			if verbose {
				fmt.Printf("found close match: %s (slug: %s)\n", mod.Name, mod.Slug)
			}
		} else {
			otherResults = append(otherResults, mod)
		}
	}
	
	// Combine results with exact matches first
	var reorderedResults []Mod
	reorderedResults = append(reorderedResults, exactMatches...)
	reorderedResults = append(reorderedResults, closeMatches...)
	reorderedResults = append(reorderedResults, otherResults...)
	
	if verbose && len(exactMatches) > 0 {
		fmt.Printf("reordered results: %d exact matches, %d close matches, %d other results\n", 
			len(exactMatches), len(closeMatches), len(otherResults))
	}
	
	return reorderedResults
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
