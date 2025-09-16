package project

// Project

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

type ContentType int
const (
	Mod ContentType = iota
	Resourcepack
	Shaderpack
	Datapack
	World
)

type HashFormat int 
const (
	SHA1 HashFormat = iota
	SHA256
	SHA512
	MD5
)

type ModSide int
const (
	None ModSide = iota
	Client
	Server
	Both
)

type Source int
const (
	Modrinth Source = iota
	Curseforge
	Custom
)

type hashes struct {
	sha1   string
	sha256 string
	sha512 string
	md5    string
}

type fileData struct {
	filename string
	filesize int64
	filepath string
	hashes   hashes
}

type DependencyType int
const (
	Required DependencyType = iota
	Optional
	Embedded
	Incompatible
)

type Dependency struct {
	name           string
	slug           string
	id             string
	dependencyType DependencyType
}

type ContentData struct {
	// todo: finish
	contentType  ContentType
	name         string
	id           string
	slug         string
	side         ModSide
	pageUrl      string
	downloadUrl  string
	versionid    string
	source       Source
	file         fileData
	dependencies []Dependency
}

type Manifest struct {
	ContentDirectory string
	Content          []ContentData
}