package homeassistant

import (
	"fmt"
	"goalfeed/config"
	"net/http"
	"os"
	"strings"
	"time"
)

// ResolveHA resolves the Home Assistant URL and token, considering Supervisor env overrides.
// Returns (url, token, source) where source is "env", "config", or "unset".
func ResolveHA() (string, string, string) {
	url := os.Getenv("SUPERVISOR_API")
	token := os.Getenv("SUPERVISOR_TOKEN")
	source := "env"
	if url == "" {
		url = config.GetString("home_assistant.url")
		source = "config"
	} else {
		url = strings.TrimRight(url, "/") + "/core"
	}
	if token == "" {
		token = config.GetString("home_assistant.access_token")
	}
	if url == "" || token == "" {
		source = "unset"
	}
	return url, token, source
}

// CheckConnection performs a lightweight authenticated request to Home Assistant.
// It returns (connected, source, message).
func CheckConnection(timeout time.Duration) (bool, string, string) {
	url, token, source := ResolveHA()
	if url == "" || token == "" {
		return false, source, "Home Assistant URL or access token not set"
	}
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", strings.TrimRight(url, "/")+"/api/", nil)
	if err != nil {
		return false, source, fmt.Sprintf("request error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return false, source, fmt.Sprintf("connection error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, source, "OK"
	}
	return false, source, fmt.Sprintf("HTTP %s", resp.Status)
}

