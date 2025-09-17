package modrinth

import "time"

// Enums for client/server side
type SideSupport int

const (
	SideRequired SideSupport = iota
	SideOptional
	SideUnsupported
	SideUnknown
)

func (s SideSupport) String() string {
	return [...]string{"required", "optional", "unsupported", "unknown"}[s]
}

// Status enums
type ProjectStatus int

const (
	StatusApproved ProjectStatus = iota
	StatusArchived
	StatusRejected
	StatusDraft
	StatusUnlisted
	StatusProcessing
	StatusWithheld
	StatusScheduled
	StatusPrivate
	StatusUnknown
)

func (s ProjectStatus) String() string {
	return [...]string{"approved", "archived", "rejected", "draft", "unlisted", "processing", "withheld", "scheduled", "private", "unknown"}[s]
}

type RequestedStatus int

const (
	ReqApproved RequestedStatus = iota
	ReqArchived
	ReqUnlisted
	ReqPrivate
	ReqDraft
)

func (r RequestedStatus) String() string {
	return [...]string{"approved", "archived", "unlisted", "private", "draft"}[r]
}

// ProjectType enums
type ProjectType int

const (
	TypeMod ProjectType = iota
	TypeModpack
	TypeResourcepack
	TypeShader
)

func (t ProjectType) String() string {
	return [...]string{"mod", "modpack", "resourcepack", "shader"}[t]
}

// MonetizationStatus enums
type MonetizationStatus int

const (
	Monetized MonetizationStatus = iota
	Demonetized
	ForceDemonetized
)

func (m MonetizationStatus) String() string {
	return [...]string{"monetized", "demonetized", "force-demonetized"}[m]
}

type ModrinthLicense struct {
	ID   string
	Name string
	URL  string
}

type ModrinthModeratorMessage struct {
	Message string
	Body    string
}

type ModrinthGalleryItem struct {
	URL         string
	Featured    bool
	Title       string
	Description string
	Created     time.Time
	Ordering    int
}

type ModrinthProject struct {
	Slug                 string
	Title                string
	Description          string
	Categories           []string
	Body                 string
	ClientSide           SideSupport
	ServerSide           SideSupport
	Status               ProjectStatus
	RequestedStatus      RequestedStatus
	AdditionalCategories []string
	IssuesURL            string
	SourceURL            string
	WikiURL              string
	DiscordURL           string
	DonationURL          string
	ProjectType          ProjectType
	Downloads            int
	Color                *int
	ThreadID             *string
	MonetizationStatus   MonetizationStatus
	ID                   string
	Team                 string
	BodyURL              *string
	ModeratorMessage     *ModrinthModeratorMessage
	Published            time.Time
	Updated              time.Time
	Approved             *time.Time
	Queued               *time.Time
	Followers            int
	License              ModrinthLicense
	Versions             []string
	GameVersions         []string
	Loaders              []string
	Gallery              []ModrinthGalleryItem
}
