package project

type ModloaderVersion struct {
	Name    string
	Version string
}

type ProjectVersions struct {
	Game   string
	Loader ModloaderVersion
}

type Project struct {
	Name          string
	Description   string
	Author        string
	Root          string
	Versions      ProjectVersions
	DefaultSource string // "modrinth" or "curseforge"
}

// Mods

type Mod struct {
	// todo: finish
}

type Manifest struct {
	ModsDirectory string
	Mods          []Mod
}