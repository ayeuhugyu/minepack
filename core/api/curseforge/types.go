package curseforge

import "time"

// curseforge API response types
type SearchResponse struct {
	Data       []Mod          `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	Index       int `json:"index"`
	PageSize    int `json:"pageSize"`
	ResultCount int `json:"resultCount"`
	TotalCount  int `json:"totalCount"`
}

type Mod struct {
	ID                   int         `json:"id"`
	GameID               int         `json:"gameId"`
	Name                 string      `json:"name"`
	Slug                 string      `json:"slug"`
	Links                ModLinks    `json:"links"`
	Summary              string      `json:"summary"`
	Status               int         `json:"status"`
	DownloadCount        int64       `json:"downloadCount"`
	IsFeatured           bool        `json:"isFeatured"`
	PrimaryCategoryID    int         `json:"primaryCategoryId"`
	Categories           []Category  `json:"categories"`
	ClassID              int         `json:"classId"`
	Authors              []ModAuthor `json:"authors"`
	Logo                 ModAsset    `json:"logo"`
	Screenshots          []ModAsset  `json:"screenshots"`
	MainFileID           int         `json:"mainFileId"`
	LatestFiles          []File      `json:"latestFiles"`
	LatestFilesIndexes   []FileIndex `json:"latestFilesIndexes"`
	DateCreated          time.Time   `json:"dateCreated"`
	DateModified         time.Time   `json:"dateModified"`
	DateReleased         time.Time   `json:"dateReleased"`
	AllowModDistribution *bool       `json:"allowModDistribution"`
	GamePopularityRank   int         `json:"gamePopularityRank"`
	IsAvailable          bool        `json:"isAvailable"`
	ThumbsUpCount        int         `json:"thumbsUpCount"`
}

type ModLinks struct {
	WebsiteURL string `json:"websiteUrl"`
	WikiURL    string `json:"wikiUrl"`
	IssuesURL  string `json:"issuesUrl"`
	SourceURL  string `json:"sourceUrl"`
}

type Category struct {
	ID               int       `json:"id"`
	GameID           int       `json:"gameId"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	URL              string    `json:"url"`
	IconURL          string    `json:"iconUrl"`
	DateModified     time.Time `json:"dateModified"`
	IsClass          bool      `json:"isClass"`
	ClassID          int       `json:"classId"`
	ParentCategoryID int       `json:"parentCategoryId"`
}

type ModAuthor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ModAsset struct {
	ID           int    `json:"id"`
	ModID        int    `json:"modId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailUrl"`
	URL          string `json:"url"`
}

type File struct {
	ID                   int                   `json:"id"`
	GameID               int                   `json:"gameId"`
	ModID                int                   `json:"modId"`
	IsAvailable          bool                  `json:"isAvailable"`
	DisplayName          string                `json:"displayName"`
	FileName             string                `json:"fileName"`
	ReleaseType          int                   `json:"releaseType"`
	FileStatus           int                   `json:"fileStatus"`
	Hashes               []FileHash            `json:"hashes"`
	FileDate             time.Time             `json:"fileDate"`
	FileLength           int64                 `json:"fileLength"`
	DownloadCount        int64                 `json:"downloadCount"`
	DownloadURL          string                `json:"downloadUrl"`
	GameVersions         []string              `json:"gameVersions"`
	SortableGameVersions []SortableGameVersion `json:"sortableGameVersions"`
	Dependencies         []FileDependency      `json:"dependencies"`
	ExposeAsAlternative  *bool                 `json:"exposeAsAlternative"`
	ParentProjectFileID  *int                  `json:"parentProjectFileId"`
	AlternateFileID      *int                  `json:"alternateFileId"`
	IsServerPack         *bool                 `json:"isServerPack"`
	ServerPackFileID     *int                  `json:"serverPackFileId"`
	FileFingerprint      int64                 `json:"fileFingerprint"`
	Modules              []FileModule          `json:"modules"`
}

type FileHash struct {
	Value string `json:"value"`
	Algo  int    `json:"algo"`
}

type SortableGameVersion struct {
	GameVersionName        string    `json:"gameVersionName"`
	GameVersionPadded      string    `json:"gameVersionPadded"`
	GameVersion            string    `json:"gameVersion"`
	GameVersionReleaseDate time.Time `json:"gameVersionReleaseDate"`
	GameVersionTypeID      int       `json:"gameVersionTypeId"`
}

type FileDependency struct {
	ModID        int `json:"modId"`
	RelationType int `json:"relationType"`
}

type FileModule struct {
	Name        string `json:"name"`
	Fingerprint int64  `json:"fingerprint"`
}

type FileIndex struct {
	GameVersion       string `json:"gameVersion"`
	FileID            int    `json:"fileId"`
	Filename          string `json:"filename"`
	ReleaseType       int    `json:"releaseType"`
	GameVersionTypeID int    `json:"gameVersionTypeId"`
	ModLoader         int    `json:"modLoader"`
}

// modloader constants
const (
	ModLoaderAny        = 0
	ModLoaderForge      = 1
	ModLoaderCauldron   = 2
	ModLoaderLiteLoader = 3
	ModLoaderFabric     = 4
	ModLoaderQuilt      = 5
	ModLoaderNeoForge   = 6
)

// game constants
const (
	GameMinecraft = 432
)

// class constants (project types)
const (
	ClassMods          = 6
	ClassModpacks      = 4471
	ClassResourcePacks = 12
	ClassShaders       = 6552
)
