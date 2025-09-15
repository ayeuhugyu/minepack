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

type hashes struct {
	sha1   string
	sha256 string
	sha512 string
	md5    string
}

type ContentData struct {
	// todo: finish
	contentType ContentType
	name        string
	side        ModSide
	hashes      hashes
}

type Manifest struct {
	ContentDirectory string
	Content          []ContentData
}