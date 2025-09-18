package curseforge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// curseforge API client with API key support
var CurseForgeClient *Client

const BaseURL = "https://api.curseforge.com/v1"

type Client struct {
	httpClient *http.Client
	apiKey     string
	userAgent  string
}

func base64decode(input string) (string, error) {
	decoded, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(input)))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	return string(decoded), nil
}

var apiKey, _ = base64decode("JDJhJDEwJGJMNGJJTDVwVVdxZmNPN0tRdG5NUmVha3d0ZkhiTktoNnYxdVRwS2x6aHdvdWVFSlFuUG5t") // please avoid using this api key if you are forking this, curseforge rate limits are stupidly low

func init() {
	CurseForgeClient = &Client{
		httpClient: &http.Client{},
		userAgent:  "Minepack/1.0 (+https://github.com/ayeuhugyu/minepack)",
		apiKey:     apiKey,
	}
}

// makes an authenticated request to the curseforge API
func (c *Client) makeRequest(endpoint string, target interface{}) error {
	url := BaseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s - Response: %s", resp.StatusCode, resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
