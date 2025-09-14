package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"goalfeed/models"

	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) (*httptest.Server, *int) {
	status := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status = status + 1 // count requests
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "token")
	return server, &status
}

func TestSendEvent_OK(t *testing.T) {
	_, count := setupTestServer(t)
	SendEvent(models.Event{Id: "e", Type: models.EventTypeGoal, LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL"})
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestSendEvent_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	SendEvent(models.Event{Id: "e", Type: models.EventTypeGoal, LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL"})
}

func TestSendGameUpdate_OK(t *testing.T) {
	_, count := setupTestServer(t)
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendGameUpdate(game)
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestSendGameUpdate_ErrorStatus(t *testing.T) {
	// server responds 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendGameUpdate(game)
}

func TestSendPeriodUpdate_OK(t *testing.T) {
	_, count := setupTestServer(t)
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendPeriodUpdate(game, models.EventTypePeriodStart)
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestSendPeriodUpdate_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendPeriodUpdate(game, models.EventTypePeriodStart)
}

func TestSendCustomEvent_OK(t *testing.T) {
	_, count := setupTestServer(t)
	SendCustomEvent("custom", map[string]interface{}{"a": 1})
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestGetLeagueName(t *testing.T) {
	if getLeagueName(models.LeagueIdNHL) != "NHL" {
		t.Fatal("nhl")
	}
	if getLeagueName(models.LeagueIdMLB) != "MLB" {
		t.Fatal("mlb")
	}
	if getLeagueName(models.LeagueIdCFL) != "CFL" {
		t.Fatal("cfl")
	}
	if getLeagueName(models.LeagueIdNFL) != "NFL" {
		t.Fatal("nfl")
	}
	if getLeagueName(999) != "Unknown" {
		t.Fatal("unknown")
	}
}

func TestCreateRichEventContainsDefaults(t *testing.T) {
	re := createRichEvent(models.Event{Type: models.EventTypeGoal})
	if re.GameState.Status != models.StatusActive {
		t.Fatalf("expected default active status in rich event")
	}
	// Ensure it is JSON marshalable
	if _, err := json.Marshal(re); err != nil {
		t.Fatalf("rich event should marshal: %v", err)
	}
}

func TestPublishBaselineForMonitoredTeams(t *testing.T) {
	// mock HA
	_, _ = setupTestServer(t)
	// configure monitored teams
	os.Setenv("GOALFEED_WATCH_NHL", "WPG")
	os.Setenv("GOALFEED_WATCH_MLB", "TOR")
	os.Setenv("GOALFEED_WATCH_CFL", "WPG")
	os.Setenv("GOALFEED_WATCH_NFL", "BUF")
	PublishBaselineForMonitoredTeams()
}

func TestSendEvent_InsideHA(t *testing.T) {
	// Test SendEvent when running inside Home Assistant add-on environment
	os.Setenv("SUPERVISOR_API", "http://supervisor")
	os.Setenv("SUPERVISOR_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("SUPERVISOR_API")
		os.Unsetenv("SUPERVISOR_TOKEN")
	}()

	event := models.Event{
		Type:        "GOAL",
		Description: "Test goal",
		TeamCode:    "WPG",
		LeagueId:    models.LeagueIdNHL,
		LeagueName:  "NHL",
		Id:          "test-id",
	}

	// Test that SendEvent runs without error
	assert.NotPanics(t, func() {
		SendEvent(event)
	})
}

func TestSendEvent_OutsideHA(t *testing.T) {
	// Test SendEvent when not running inside Home Assistant
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")

	event := models.Event{
		Type:        "GOAL",
		Description: "Test goal",
		TeamCode:    "WPG",
		LeagueId:    models.LeagueIdNHL,
		LeagueName:  "NHL",
		Id:          "test-id",
	}

	// Test that SendEvent runs without error
	assert.NotPanics(t, func() {
		SendEvent(event)
	})
}

func TestSendEvent_EmptyEvent(t *testing.T) {
	// Test SendEvent with empty event
	event := models.Event{}

	// Test that SendEvent runs without error
	assert.NotPanics(t, func() {
		SendEvent(event)
	})
}
