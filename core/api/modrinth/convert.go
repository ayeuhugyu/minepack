package modrinth

import (
	"fmt"
	"minepack/core/project"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
)

func modrinthProjectTypeToContentType(pt string) project.ContentType {
	switch pt {
	case "mod":
		return project.Mod
	case "resourcepack":
		return project.Resourcepack
	case "shader":
		return project.Shaderpack
	default:
		return project.Mod
	}
}

func modrinthProjectTypeToFileLocationPrefix(pt string) string {
	switch pt {
	case "mod":
		return "mods/"
	case "resourcepack":
		return "resourcepacks/"
	case "shader":
		return "shaderpacks/"
	default:
		return "mods/"
	}
}

func modrinthProjectSideToModSideData(pt string) project.ModSideData {
	switch pt {
	case "required":
		return project.SideRequired
	case "optional":
		return project.SideOptional
	case "unsupported":
		return project.SideUnsupported
	case "unknown":
		return project.SideUnknown
	default:
		return project.SideRequired
	}
}

// converts a full project to ContentData with version info
func ConvertProjectToContentData(proj *modrinth.Project, version *modrinth.Version) project.ContentData {
	// handle pointer dereferences safely
	name := ""
	if proj.Title != nil {
		name = *proj.Title
	}

	id := ""
	if proj.ID != nil {
		id = *proj.ID
	}

	slug := ""
	if proj.Slug != nil {
		slug = *proj.Slug
	}

	contentData := project.ContentData{
		ContentType: modrinthProjectTypeToContentType(*proj.ProjectType),
		Name:        name,
		Id:          id,
		Slug:        slug,
		Side: project.ModSide{
			Client: modrinthProjectSideToModSideData(*proj.ClientSide),
			Server: modrinthProjectSideToModSideData(*proj.ServerSide),
		},
		PageUrl:      fmt.Sprintf("https://modrinth.com/mod/%s", slug),
		Source:       project.Modrinth,
		Dependencies: []project.Dependency{},
	}

	// set content type
	if proj.ProjectType != nil {
		switch *proj.ProjectType {
		case "mod":
			contentData.ContentType = project.Mod
		case "resourcepack":
			contentData.ContentType = project.Resourcepack
		case "shader":
			contentData.ContentType = project.Shaderpack
		}
	}

	// add version and file information if available
	if version != nil {
		if version.ID != nil {
			contentData.VersionId = *version.ID
		}
		if len(version.Files) > 0 {
			file := version.Files[0]
			if file.URL != nil {
				contentData.DownloadUrl = *file.URL
			}

			filename := ""
			if file.Filename != nil {
				filename = *file.Filename
			}

			size := int64(0)
			if file.Size != nil {
				size = int64(*file.Size)
			}

			contentData.File = project.FileData{
				Filename: filename,
				Filesize: size,
				Filepath: modrinthProjectTypeToFileLocationPrefix(*proj.ProjectType) + filename,
			}

			// add hashes if available
			if file.Hashes != nil {
				hashes := project.Hashes{}
				if sha1, ok := file.Hashes["sha1"]; ok {
					hashes.Sha1 = sha1
				}
				if sha256, ok := file.Hashes["sha256"]; ok {
					hashes.Sha256 = sha256
				}
				if sha512, ok := file.Hashes["sha512"]; ok {
					hashes.Sha512 = sha512
				}
				contentData.File.Hashes = hashes
			}
		}

		// add dependencies if available
		if len(version.Dependencies) > 0 {
			for _, dep := range version.Dependencies {
				// fetch the dependency project to get its slug and name
				depProj, err := GetProject(*dep.ProjectID)
				if err != nil {
					continue
				}
				depName := ""
				if depProj.Title != nil {
					depName = *depProj.Title
				}
				depSlug := ""
				if depProj.Slug != nil {
					depSlug = *depProj.Slug
				}

				depType := project.Required
				if dep.DependencyType != nil {
					switch *dep.DependencyType {
					case "required":
						depType = project.Required
					case "optional":
						depType = project.Optional
					case "incompatible":
						depType = project.Incompatible
					case "embedded":
						depType = project.Embedded
					}
				}

				contentData.Dependencies = append(contentData.Dependencies, project.Dependency{
					Name:           depName,
					Slug:           depSlug,
					Id:             *dep.ProjectID,
					DependencyType: depType,
				})
			}
		}
	}

	return contentData
}
