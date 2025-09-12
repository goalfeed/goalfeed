package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/targets/applog"
	"goalfeed/utils"
	"net/http"
	"os"
	"time"
)

var logger = utils.GetLogger()

// SendEvent sends a detailed event to Home Assistant
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

	// Create a rich event with additional context
	richEvent := createRichEvent(event)

	// Convert the event to JSON
	jsonData, err := json.Marshal(richEvent)
	if err != nil {
		logger.Error("Failed to marshal event: " + err.Error())
		return
	}

	// Create a new request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create request: " + err.Error())
		return
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
		ok := false
		applog.Append(models.AppLogEntry{Type: models.AppLogTypeEvent, LeagueId: models.League(event.LeagueId), LeagueName: event.LeagueName, TeamCode: event.TeamCode, Opponent: event.OpponentCode, GameCode: event.GameCode, Event: &event, Target: "ha:event:goal", Success: &ok, Error: err.Error(), CorrelationId: event.Id})
		return
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn(resp.Status)
		logger.Warn("Failed to send event to Home Assistant")
		ok := false
		applog.Append(models.AppLogEntry{Type: models.AppLogTypeEvent, LeagueId: models.League(event.LeagueId), LeagueName: event.LeagueName, TeamCode: event.TeamCode, Opponent: event.OpponentCode, GameCode: event.GameCode, Event: &event, Target: "ha:event:goal", Success: &ok, Error: resp.Status, CorrelationId: event.Id})
	} else {
		logger.Info(fmt.Sprintf("Successfully sent %s event to Home Assistant", event.Type))
		ok := true
		applog.Append(models.AppLogEntry{Type: models.AppLogTypeEvent, LeagueId: models.League(event.LeagueId), LeagueName: event.LeagueName, TeamCode: event.TeamCode, Opponent: event.OpponentCode, GameCode: event.GameCode, Event: &event, Target: "ha:event:goal", Success: &ok, CorrelationId: event.Id})
	}
}

// SendGameUpdate sends detailed game state updates to Home Assistant
func SendGameUpdate(game models.Game) {
	homeAssistantURL := os.Getenv("SUPERVISOR_API")
	accessToken := os.Getenv("SUPERVISOR_TOKEN")

	if homeAssistantURL == "" {
		homeAssistantURL = config.GetString("home_assistant.url")
	} else {
		homeAssistantURL = homeAssistantURL + "/core"
	}
	if accessToken == "" {
		accessToken = config.GetString("home_assistant.access_token")
	}

	url := homeAssistantURL + "/api/events/game_update"

	// Create a comprehensive game update payload
	gameUpdate := map[string]interface{}{
		"gameCode":      game.GameCode,
		"leagueId":      game.LeagueId,
		"leagueName":    getLeagueName(game.LeagueId),
		"status":        game.CurrentState.Status,
		"period":        game.CurrentState.Period,
		"periodType":    game.CurrentState.PeriodType,
		"timeRemaining": game.CurrentState.TimeRemaining,
		"clock":         game.CurrentState.Clock,
		"homeTeam": map[string]interface{}{
			"code":         game.CurrentState.Home.Team.TeamCode,
			"name":         game.CurrentState.Home.Team.TeamName,
			"score":        game.CurrentState.Home.Score,
			"periodScores": game.CurrentState.Home.PeriodScores,
		},
		"awayTeam": map[string]interface{}{
			"code":         game.CurrentState.Away.Team.TeamCode,
			"name":         game.CurrentState.Away.Team.TeamName,
			"score":        game.CurrentState.Away.Score,
			"periodScores": game.CurrentState.Away.PeriodScores,
		},
		"venue":      game.CurrentState.Venue,
		"weather":    game.CurrentState.Weather,
		"statistics": game.Statistics,
		"timestamp":  time.Now().Format(time.RFC3339),
		"isFetching": game.IsFetching,
	}

	jsonData, err := json.Marshal(gameUpdate)
	if err != nil {
		logger.Error("Failed to marshal game update: " + err.Error())
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create game update request: " + err.Error())
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Failed to send game update to Home Assistant: " + err.Error())
		ok := false
		applog.Append(models.AppLogEntry{Type: models.AppLogTypeStateChange, LeagueId: game.LeagueId, LeagueName: getLeagueName(game.LeagueId), TeamCode: game.CurrentState.Home.Team.TeamCode, Opponent: game.CurrentState.Away.Team.TeamCode, GameCode: game.GameCode, Metric: "game_update", Success: &ok, Error: err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn("Failed to send game update to Home Assistant: " + resp.Status)
		ok := false
		applog.Append(models.AppLogEntry{Type: models.AppLogTypeStateChange, LeagueId: game.LeagueId, LeagueName: getLeagueName(game.LeagueId), TeamCode: game.CurrentState.Home.Team.TeamCode, Opponent: game.CurrentState.Away.Team.TeamCode, GameCode: game.GameCode, Metric: "game_update", Success: &ok, Error: resp.Status})
	}

	// Removed duplicate team.current_score state logging here; sensor publishing handles change logs
}

// SendPeriodUpdate sends period/quarter start/end events
func SendPeriodUpdate(game models.Game, eventType models.EventType) {
	homeAssistantURL := os.Getenv("SUPERVISOR_API")
	accessToken := os.Getenv("SUPERVISOR_TOKEN")

	if homeAssistantURL == "" {
		homeAssistantURL = config.GetString("home_assistant.url")
	} else {
		homeAssistantURL = homeAssistantURL + "/core"
	}
	if accessToken == "" {
		accessToken = config.GetString("home_assistant.access_token")
	}

	url := homeAssistantURL + "/api/events/period_update"

	periodUpdate := map[string]interface{}{
		"eventType":  eventType,
		"gameCode":   game.GameCode,
		"leagueId":   game.LeagueId,
		"leagueName": getLeagueName(game.LeagueId),
		"period":     game.CurrentState.Period,
		"periodType": game.CurrentState.PeriodType,
		"homeTeam":   game.CurrentState.Home.Team.TeamName,
		"awayTeam":   game.CurrentState.Away.Team.TeamName,
		"homeScore":  game.CurrentState.Home.Score,
		"awayScore":  game.CurrentState.Away.Score,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(periodUpdate)
	if err != nil {
		logger.Error("Failed to marshal period update: " + err.Error())
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create period update request: " + err.Error())
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Failed to send period update to Home Assistant: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn("Failed to send period update to Home Assistant: " + resp.Status)
	}
}

// createRichEvent creates a RichEvent with additional context for Home Assistant
func createRichEvent(event models.Event) models.RichEvent {
	return models.RichEvent{
		Event: event,
		// Add additional context that would be useful for Home Assistant
		GameState: models.GameState{
			// This would be populated with current game state
			Status: models.StatusActive, // Default, would be updated with actual state
		},
		Weather: models.Weather{
			// Weather info if available
		},
		BroadcastInfo: models.BroadcastInfo{
			Networks:     []string{},
			Language:     "en",
			Availability: "free",
		},
	}
}

// getLeagueName returns the league name for a given league ID
func getLeagueName(leagueId models.League) string {
	switch leagueId {
	case models.LeagueIdNHL:
		return "NHL"
	case models.LeagueIdMLB:
		return "MLB"
	case models.LeagueIdEPL:
		return "EPL"
	case models.LeagueIdIIHF:
		return "IIHF"
	case models.LeagueIdCFL:
		return "CFL"
	default:
		return "Unknown"
	}
}

// SendCustomEvent sends a custom event to Home Assistant with full control over the payload
func SendCustomEvent(eventType string, data map[string]interface{}) {
	homeAssistantURL := os.Getenv("SUPERVISOR_API")
	accessToken := os.Getenv("SUPERVISOR_TOKEN")

	if homeAssistantURL == "" {
		homeAssistantURL = config.GetString("home_assistant.url")
	} else {
		homeAssistantURL = homeAssistantURL + "/core"
	}
	if accessToken == "" {
		accessToken = config.GetString("home_assistant.access_token")
	}

	url := fmt.Sprintf("%s/api/events/%s", homeAssistantURL, eventType)

	// Add timestamp and source
	data["timestamp"] = time.Now().Format(time.RFC3339)
	data["source"] = "goalfeed"

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal custom event: " + err.Error())
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create custom event request: " + err.Error())
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Failed to send custom event to Home Assistant: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn("Failed to send custom event to Home Assistant: " + resp.Status)
	}
}
