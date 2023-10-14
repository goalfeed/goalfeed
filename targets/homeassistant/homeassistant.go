package homeassistant

import (
	"bytes"
	"encoding/json"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/utils"
	"net/http"
	"os"
)

var logger = utils.GetLogger()

func SendEvent(event models.Event) {
	// Detect if running inside Home Assistant add-on environment
	homeAssistantURL := os.Getenv("SUPERVISOR_API")
	accessToken := os.Getenv("SUPERVISOR_TOKEN")

	// If not running inside Home Assistant, use the existing configuration
	if homeAssistantURL == "" {
		homeAssistantURL = config.GetString("home_assistant.url")
	} else {
		homeAssistantURL = homeAssistantURL + "/core"
	}
	if accessToken == "" {
		accessToken = config.GetString("home_assistant.access_token")
	}

	// Construct the URL for the Home Assistant event endpoint
	url := homeAssistantURL + "/api/events/goal"

	// Convert the event to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic("Failed to create request")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn(err)
		logger.Warn("Failed to send event to Home Assistant")
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes (you can expand on this as needed)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn(resp.Status)
		logger.Warn("Failed to send event to Home Assistant")
	}
}
