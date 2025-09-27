package curseforge

import (
	"fmt"
	"minepack/core/project"
)

// converts curseforge Mod to ContentData
func ConvertToContentData(mod *Mod) project.ContentData {
	contentData := project.ContentData{
		ContentType: getContentType(mod.ClassID),
		Name:        mod.Name,
		Id:          fmt.Sprintf("%d", mod.ID),
		Slug:        mod.Slug,
		Side: project.ModSide{
			Server: project.SideRequired,
			Client: project.SideRequired,
		}, // default, curseforge doesn't have explicit client/server side info in search
		PageUrl:      fmt.Sprintf("https://www.curseforge.com/minecraft/mc-mods/%s", mod.Slug),
		Source:       project.Curseforge,
		Dependencies: []project.Dependency{},
	}

	// add download URL and file info if available
	if len(mod.LatestFiles) > 0 {
		latestFile := mod.LatestFiles[0]
		contentData.DownloadUrl = latestFile.DownloadURL
		contentData.VersionId = fmt.Sprintf("%d", latestFile.ID)

		contentData.File = project.FileData{
			Filename: latestFile.FileName,
			Filesize: latestFile.FileLength,
			Filepath: latestFile.FileName,
		}

		// convert hashes
		if len(latestFile.Hashes) > 0 {
			hashes := project.Hashes{}
			for _, hash := range latestFile.Hashes {
				switch hash.Algo {
				case 1: // SHA1
					hashes.Sha1 = hash.Value
				case 2: // MD5
					hashes.Md5 = hash.Value
				}
			}
			contentData.File.Hashes = hashes
		}
	}

	return contentData
}

// converts a full Mod with specific file to ContentData
func ConvertModToContentData(mod *Mod, file *File) project.ContentData {
	contentData := project.ContentData{
		ContentType: getContentType(mod.ClassID),
		Name:        mod.Name,
		Id:          fmt.Sprintf("%d", mod.ID),
		Slug:        mod.Slug,
		Side: project.ModSide{
			Server: project.SideRequired,
			Client: project.SideRequired,
		},
		PageUrl:      fmt.Sprintf("https://www.curseforge.com/minecraft/mc-mods/%s", mod.Slug),
		Source:       project.Curseforge,
		Dependencies: []project.Dependency{},
	}

	// add file information if provided
	if file != nil {
		contentData.DownloadUrl = file.DownloadURL
		contentData.VersionId = fmt.Sprintf("%d", file.ID)

		contentData.File = project.FileData{
			Filename: file.FileName,
			Filesize: file.FileLength,
			Filepath: file.FileName,
		}

		// convert hashes
		if len(file.Hashes) > 0 {
			hashes := project.Hashes{}
			for _, hash := range file.Hashes {
				switch hash.Algo {
				case 1: // SHA1
					hashes.Sha1 = hash.Value
				case 2: // MD5
					hashes.Md5 = hash.Value
				}
			}
			contentData.File.Hashes = hashes
		}

		// convert dependencies
		for _, dep := range file.Dependencies {
			var depType project.DependencyType
			switch dep.RelationType {
			case 1: // EmbeddedLibrary
				depType = project.Embedded
			case 2: // OptionalDependency
				depType = project.Optional
			case 3: // RequiredDependency
				depType = project.Required
			case 4: // Tool
				depType = project.Optional
			case 5: // Incompatible
				depType = project.Incompatible
			case 6: // Include
				depType = project.Required
			default:
				depType = project.Optional
			}

			modData, err := GetProject(fmt.Sprintf("%d", dep.ModID))
			if err != nil {
				fmt.Printf("error fetching dependency data for mod ID %d: %v\n", dep.ModID, err)
			}

			contentData.Dependencies = append(contentData.Dependencies, project.Dependency{
				Id:             fmt.Sprintf("%d", dep.ModID),
				DependencyType: depType,
				Name:           modData.Name,
				Slug:           modData.Slug,
			})
		}
	}

	return contentData
}

// getContentType converts CurseForge class ID to our ContentType
func getContentType(classID int) project.ContentType {
	switch classID {
	case ClassMods:
		return project.Mod
	case ClassModpacks:
		return project.Mod // We don't have a specific modpack type, use Mod
	case ClassResourcePacks:
		return project.Resourcepack
	case ClassShaders:
		return project.Shaderpack
	default:
		return project.Mod // Default fallback
	}
}
