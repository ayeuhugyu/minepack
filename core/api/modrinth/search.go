package modrinth

import (
	"fmt"
	"minepack/core/project"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
)

// searches modrinth for projects matching the query and optional pack data
func SearchProjects(query string, projectData project.Project, verbose bool) ([]*modrinth.SearchResult, error) {
	if verbose {
		fmt.Printf("starting search for query: %s\n", query)
		if projectData.Versions.Game != "" || projectData.Versions.Loader.Name != "" {
			fmt.Printf("pack data provided: gameVersion=%s, modloader=%s\n", projectData.Versions.Game, projectData.Versions.Loader.Name)
		}
	}

	searchOptions := &modrinth.SearchOptions{
		Query: query,
		Limit: 10,
	}

	// add facets if project data is provided
	if projectData != (project.Project{}) {
		var facets [][]string

		// add game version facet
		if projectData.Versions.Game != "" {
			facets = append(facets, []string{fmt.Sprintf("versions:%s", projectData.Versions.Game)})
		}

		// add project types facet
		facets = append(facets, []string{"project_type:mod", "project_type:resourcepack", "project_type:shader"})

		// add modloader facet
		if projectData.Versions.Loader.Name != "" {
			facets = append(facets, []string{fmt.Sprintf("categories:%s", projectData.Versions.Loader.Name)})
		}

		searchOptions.Facets = facets

		if verbose {
			fmt.Printf("constructed facets: %+v\n", facets)
		}
	}

	response, err := ModrinthClient.Projects.Search(searchOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	if verbose {
		fmt.Printf("search completed: %d results found\n", len(response.Hits))
	}

	return response.Hits, nil
}
