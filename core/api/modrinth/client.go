package modrinth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

const baseURL = "https://api.modrinth.com/v2"

var httpClient = &http.Client{}

func getJSON(url string, target any) error {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Minepack/1.0 (+https://github.com/ayeuhugyu/minepack)")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	// print the raw http request for debugging --- IGNORE ---
	dump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		fmt.Printf("%q\n", dump)
	}
	// --- IGNORE ---
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
