package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"goalfeed/models"
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

func TestSendGameUpdate_OK(t *testing.T) {
	_, count := setupTestServer(t)
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendGameUpdate(game)
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestSendPeriodUpdate_OK(t *testing.T) {
	_, count := setupTestServer(t)
	game := models.Game{GameCode: "g1", LeagueId: models.LeagueIdNHL}
	SendPeriodUpdate(game, models.EventTypePeriodStart)
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
	}
}

func TestSendCustomEvent_OK(t *testing.T) {
	_, count := setupTestServer(t)
	SendCustomEvent("custom", map[string]interface{}{"a": 1})
	if *count == 0 {
		t.Fatalf("expected at least one request sent")
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
