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

type Hashes struct {
	Sha1   string
	Sha256 string
	Sha512 string
	Md5    string
}

type FileData struct {
	Filename string
	Filesize int64
	Filepath string // relative to project root
	Hashes   Hashes
}

type DependencyType int
const (
	Required DependencyType = iota
	Optional
	Embedded
	Incompatible
)

type Dependency struct {
	Name           string
	Slug           string
	Id             string
	DependencyType DependencyType
}

type ContentData struct {
	// todo: finish
	ContentType  ContentType
	Name         string
	Id           string
	Slug         string
	Side         ModSide
	PageUrl      string
	DownloadUrl  string
	VersionId    string
	Source       Source
	File         FileData
	Dependencies []Dependency
}

type Manifest struct {
	ContentDirectory string
	Content          []ContentData
}