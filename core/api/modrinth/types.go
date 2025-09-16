package modrinth

// ModrinthContent represents the structure of a Modrinth project response (simplified)
type ModrinthContent struct {
	ID          string
	Slug        string
	Title       string
	Description string
	Author      string
	ProjectType string
	Versions    []ModrinthVersion
	IconURL     string
	PageURL     string
}

type ModrinthVersion struct {
	ID         string
	Name       string
	Version    string
	Files      []ModrinthFile
	Loaders    []string
	GameVersions []string
}

type ModrinthFile struct {
	URL      string
	Filename string
	Size     int64
	Hashes   map[string]string // e.g. sha1, sha512, etc.
}

type ModrinthSearchResult struct {
	ID    string
	Slug  string
	Title string
	Author string
	ProjectType string
	IconURL string
	Description string
}
