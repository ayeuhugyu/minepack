package modrinth

import (
	"fmt"
)

// SearchMods searches Modrinth for mods matching the query and returns results
func SearchMods(query string, loaders []string, gameVersions []string) ([]ModrinthSearchResult, error) {
	url := fmt.Sprintf("%s/search?query=%s", baseURL, query)
	// TODO: add loader/game version filters to URL
	var results []ModrinthSearchResult
	if err := getJSON(url, &results); err != nil {
		return nil, err
	}
	return results, nil
}
