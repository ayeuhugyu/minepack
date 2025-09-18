package curseforge

import (
	"fmt"
	"minepack/core/project"
)

// fetches detailed project information from CurseForge
func GetProject(projectID string) (*Mod, error) {
	endpoint := "/mods/" + projectID

	var response struct {
		Data Mod `json:"data"`
	}

	if err := CurseForgeClient.makeRequest(endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get curseforge project %s: %w", projectID, err)
	}

	return &response.Data, nil
}

// fetches all files for a project from CurseForge
func GetProjectFiles(projectID string, packData project.Project) ([]File, error) {
	endpoint := "/mods/" + projectID + "/files"
	params := ""
	if packData.Versions.Game != "" {
		params += "?gameVersion=" + packData.Versions.Game
	}
	if packData.Versions.Loader.Name != "" {
		var loaderType int = getModLoaderID(packData.Versions.Loader.Name)
		if params == "" {
			params += "?modLoaderType=" + fmt.Sprintf("%d", loaderType)
		} else {
			params += "&modLoaderType=" + fmt.Sprintf("%d", loaderType)
		}
	}

	var response struct {
		Data []File `json:"data"`
	}

	if err := CurseForgeClient.makeRequest(endpoint+params, &response); err != nil {
		return nil, fmt.Errorf("failed to get files for curseforge project %s: %w", projectID, err)
	}

	return response.Data, nil
}

// fetches filtered versions for a project (matching modrinth signature)
func GetProjectVersions(projectID string, packData project.Project) ([]File, error) {
	// get all files first
	files, err := GetProjectFiles(projectID, packData)
	if err != nil {
		return nil, err
	}

	// files have already been filtered based on pack data, it would be too complicated to verify again here (especially verifying modloader support)

	return files, nil
}
