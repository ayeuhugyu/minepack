package modrinth

import (
	"fmt"
	"minepack/core/project"
)

// DownloadModrinthContent downloads a Modrinth file to the given location
func DownloadModrinthContent(data project.ContentData, destPath string) error {
	if data.Source != project.Modrinth {
		return fmt.Errorf("content source is not modrinth")
	}
	return downloadFile(data.DownloadUrl, destPath)
}
