package modrinth

import (
	"fmt"
	"minepack/core/project"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
)

// fetches detailed project information
func GetProject(projectID string) (*modrinth.Project, error) {
	project, err := ModrinthClient.Projects.Get(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectID, err)
	}
	return project, nil
}

// fetches all versions for a project
func GetProjectVersions(projectID string, packData project.Project) ([]*modrinth.Version, error) {
	versions, err := ModrinthClient.Versions.ListVersions(projectID, modrinth.ListVersionsOptions{
		GameVersions: []string{packData.Versions.Game},
		Loaders:      []string{packData.Versions.Loader.Name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get versions for project %s: %w", projectID, err)
	}
	return versions, nil
}
