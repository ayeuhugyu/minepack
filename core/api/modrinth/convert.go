package modrinth

import (
	"fmt"
	"minepack/core/project"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
)

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
		ContentType:  project.Mod,
		Name:         name,
		Id:           id,
		Slug:         slug,
		Side:         project.Both,
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
				Filepath: filename,
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
	}

	return contentData
}
