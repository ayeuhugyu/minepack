package modrinth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "https://api.modrinth.com/v2"

func getJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("modrinth api error: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := createFile(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// createFile is a helper to create a file (can be replaced with os.Create)
func createFile(filepath string) (io.WriteCloser, error) {
	return os.Create(filepath)
}
