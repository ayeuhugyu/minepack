package modrinth

import (
	"fmt"
	"minepack/core/project"
)

// GetModInfo fetches mod info from Modrinth and returns ModrinthContent
func GetModInfo(projectID string) (*ModrinthContent, error) {
	url := fmt.Sprintf("%s/project/%s", baseURL, projectID)
	var mod ModrinthContent
	if err := getJSON(url, &mod); err != nil {
		return nil, err
	}
	// Fetch versions
	verURL := fmt.Sprintf("%s/project/%s/versions", baseURL, projectID)
	var versions []ModrinthVersion
	if err := getJSON(verURL, &versions); err == nil {
		mod.Versions = versions
	}
	return &mod, nil
}

// ConvertModrinthToContentData converts ModrinthContent to ContentData
func ConvertModrinthToContentData(mod *ModrinthContent) project.ContentData {
	var file project.FileData
	if len(mod.Versions) > 0 && len(mod.Versions[0].Files) > 0 {
		f := mod.Versions[0].Files[0]
		file = project.FileData{
			Filename: f.Filename,
			Filesize: f.Size,
			Filepath: f.Filename,
			Hashes: project.Hashes{
				Sha1:   f.Hashes["sha1"],
				Sha256: f.Hashes["sha256"],
				Sha512: f.Hashes["sha512"],
				Md5:    f.Hashes["md5"],
			},
		}
	}
	return project.ContentData{
		ContentType:  project.Mod,
		Name:         mod.Title,
		Id:           mod.ID,
		Slug:         mod.Slug,
		Side:         project.Both,
		PageUrl:      mod.PageURL,
		DownloadUrl:  file.Filepath,
		VersionId:    mod.Versions[0].ID,
		Source:       project.Modrinth,
		File:         file,
		Dependencies: []project.Dependency{},
	}
}
