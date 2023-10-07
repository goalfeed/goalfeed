package homeassistant

import (
	"bytes"
	"encoding/json"
	"goalfeed/config"
	"goalfeed/models"
	"net/http"
)

func SendEvent(event models.Event) {
	homeAssistantURL := config.GetString("home_assistant.url")
	accessToken := config.GetString("home_assistant.access_token")

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
		panic("Failed to send event to Home Assistant")
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes (you can expand on this as needed)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		panic("Failed to send event to Home Assistant")
	}
}
