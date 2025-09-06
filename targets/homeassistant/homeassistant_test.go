package homeassistant

import (
	"goalfeed/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestSendEvent(t *testing.T) {
	// Mock Home Assistant server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected method POST; got %s", r.Method)
		}
		if r.URL.String() != "/core/api/events/goal" {
			t.Errorf("Expected URL /core/api/events/goal; got %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set environment variables to mock running inside Home Assistant add-on environment
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "test-token")

	// Create a test event
	event := models.Event{
		TeamCode:     "TEST",
		TeamName:     "Test Team",
		TeamHash:     "testhash",
		LeagueId:     1,
		LeagueName:   "NHL",
		OpponentCode: "OPP",
		OpponentName: "Opponent",
		OpponentHash: "opphash",
	}

	// Call SendEvent
	SendEvent(event)

	// Reset environment variables (cleanup)
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
}

func TestSendEventWithConfig(t *testing.T) {
	// Reset viper
	viper.Reset()
	
	// Mock Home Assistant server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected method POST; got %s", r.Method)
		}
		if r.URL.String() != "/api/events/goal" {
			t.Errorf("Expected URL /api/events/goal; got %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set configuration values instead of environment variables
	viper.Set("home_assistant.url", server.URL)
	viper.Set("home_assistant.access_token", "config-token")

	// Create a test event
	event := models.Event{
		TeamCode:     "TEST",
		TeamName:     "Test Team",
		TeamHash:     "testhash",
		LeagueId:     1,
		LeagueName:   "NHL",
		OpponentCode: "OPP",
		OpponentName: "Opponent",
		OpponentHash: "opphash",
	}

	// Call SendEvent
	SendEvent(event)
}

func TestSendEventServerError(t *testing.T) {
	// Mock Home Assistant server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Set environment variables
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "test-token")
	defer os.Unsetenv("SUPERVISOR_API")
	defer os.Unsetenv("SUPERVISOR_TOKEN")

	// Create a test event
	event := models.Event{
		TeamCode:     "TEST",
		TeamName:     "Test Team",
		TeamHash:     "testhash",
		LeagueId:     1,
		LeagueName:   "NHL",
		OpponentCode: "OPP",
		OpponentName: "Opponent",
		OpponentHash: "opphash",
	}

	// Call SendEvent - should not panic
	SendEvent(event)
}
