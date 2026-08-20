package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goalfeed/models"
	"goalfeed/targets/applog"

	"github.com/spf13/viper"
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
	if getLeagueName(models.LeagueIdOlympicMensHockey) != "Olympic Men's Hockey" {
		t.Fatal("olympic men's hockey")
	}
	if getLeagueName(models.LeagueIdOlympicWomensHockey) != "Olympic Women's Hockey" {
		t.Fatal("olympic women's hockey")
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

// TestSendEvent_AlwaysPostsToGoalEndpoint pins down the README's documented
// guarantee: every event SendEvent transmits - real goal detections,
// period-start notices, and synthetic test goals - is posted to the same HA
// REST endpoint, /api/events/goal, regardless of the event's own Type field.
// Home Assistant automations must therefore filter on teamCode/leagueId, not
// on the HA event type, since there is only ever one HA event type: "goal".
func TestSendEvent_AlwaysPostsToGoalEndpoint(t *testing.T) {
	var gotPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	os.Setenv("SUPERVISOR_API", server.URL)
	os.Setenv("SUPERVISOR_TOKEN", "token")
	defer func() {
		os.Unsetenv("SUPERVISOR_API")
		os.Unsetenv("SUPERVISOR_TOKEN")
	}()

	// A real goal detection.
	SendEvent(models.Event{Id: "g1", Type: models.EventTypeGoal, TeamCode: "WPG", LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL"})
	// A period-start notice, as fired by main.go's checkGame on a period change.
	SendEvent(models.Event{Id: "p1", Type: models.EventTypePeriodStart, LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL", Description: "Period 2 started"})
	// A synthetic test goal, as fired by main.go's sendTestGoal (Type left unset).
	SendEvent(models.Event{TeamCode: "TEST", TeamName: "TEST", TeamHash: "TESTTEST", LeagueName: "TEST"})

	if len(gotPaths) != 3 {
		t.Fatalf("expected 3 requests to be sent, got %d", len(gotPaths))
	}
	for i, p := range gotPaths {
		if !strings.HasSuffix(p, "/api/events/goal") {
			t.Fatalf("request %d: expected every SendEvent call to hit .../api/events/goal regardless of event type, got %q", i, p)
		}
	}
}

// TestHomeAssistantTarget_RefusesToSendTokenToUnvalidatedURL is the important
// defense-in-depth guarantee: even if home_assistant.url has been poisoned
// (via the API, an env var, or a hand-edited config.yaml) to point at a
// public, non-local address, SendEvent must refuse to attach the Home
// Assistant access token to an outbound request rather than sending it.
//
// It uses 192.0.2.1 (TEST-NET-1, RFC 5737 - reserved for documentation and
// guaranteed non-routable) as the "attacker" host. The outbound http.Client
// used by SendEvent has no timeout set, so if the URL guard ever regresses,
// a real connection attempt to that address would hang far longer than this
// test's assertion window; bounding the call in a goroutine with a short
// timeout turns "the guard silently stopped working" into a fast, clear
// failure instead of a multi-minute hang.
func TestHomeAssistantTarget_RefusesToSendTokenToUnvalidatedURL(t *testing.T) {
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	viper.Set("home_assistant.url", "http://192.0.2.1")
	viper.Set("home_assistant.access_token", "super-secret-token")
	viper.Set("home_assistant.allow_remote_url", false)
	defer func() {
		viper.Set("home_assistant.url", "")
		viper.Set("home_assistant.access_token", "")
	}()

	logPath := filepath.Join(t.TempDir(), "app.log.jsonl")
	applog.SetLogFilePathForTest(logPath)

	done := make(chan struct{})
	go func() {
		SendEvent(models.Event{Id: "guard-test", Type: models.EventTypeGoal, LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("SendEvent did not return quickly; it appears to have attempted a real network request instead of refusing an unvalidated url")
	}

	entries := applog.Query(int(models.LeagueIdNHL), "", time.Time{}, 50)
	var found bool
	for _, e := range entries {
		if e.CorrelationId != "guard-test" {
			continue
		}
		found = true
		if e.Success == nil || *e.Success {
			t.Fatalf("expected the refused send to be logged as a failure, got success=%v", e.Success)
		}
	}
	if !found {
		t.Fatal("expected a logged failure entry for the refused send")
	}
}
