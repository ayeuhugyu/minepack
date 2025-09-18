package curseforge

import (
	"fmt"
	"io"
	"minepack/core/project"
	"net/http"
	"os"
)

// downloads a file from a ContentData entry
func DownloadContent(data project.ContentData, destPath string) error {
	if data.Source != project.Curseforge {
		return fmt.Errorf("content source is not curseforge")
	}

	if data.DownloadUrl == "" {
		return fmt.Errorf("no download URL available")
	}

	// download the file using standard HTTP client
	resp, err := http.Get(data.DownloadUrl)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// create the destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
