package modrinth

import (
	"fmt"
	"minepack/core/project"
	// "net/url"
)

// SearchProjects searches modrinth for projects matching the query and optional pack data
func SearchProjects(query string, projectData project.Project, verbose bool) ([]*ModrinthProject, error) {
	if verbose {
		fmt.Printf("Starting search for query: %s\n", query)
		if projectData.Versions.Game != "" || projectData.Versions.Loader.Name != "" {
			fmt.Printf("Pack data provided: gameVersion=%s, modloader=%s\n", projectData.Versions.Game, projectData.Versions.Loader.Name)
		}
	}
	var facets string
	if projectData != (project.Project{}) {
		facets = `&facets=[[\"versions:` + projectData.Versions.Game + `\"],[\"project_type:mod\",\"project_type:resourcepack\",\"project_type:shader\"],[\"categories:` + projectData.Versions.Loader.Name + `\"]]`
        if verbose {
            fmt.Printf("Constructed facets: %s\n", facets)
        }
    }
    url := fmt.Sprintf("%s/search?query=%s%s&limit=10", baseURL, query, facets)
    if verbose {
        fmt.Printf("Final search URL: %s\n", url)
    }
    var response *ModrinthSearchResponse = &ModrinthSearchResponse{}
    err := getJSON(url, response)
    if err != nil {
        return nil, err
    }
    return response.Hits, nil
}