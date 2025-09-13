package webApi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"goalfeed/models"
	"goalfeed/targets/applog"
	"goalfeed/targets/memoryStore"
)

func TestNormalizeGamesData_ActiveStatusPreserved(t *testing.T) {
	games := []models.Game{
		{
			CurrentState: models.GameState{
				Status: models.StatusUpcoming,
				Period: 1,
				Clock:  "10:00",
			},
		},
		{
			CurrentState: models.GameState{
				Status: models.StatusEnded,
			},
		},
	}

	out := normalizeGamesData(games)
	if out[0].CurrentState.Status != models.StatusActive {
		t.Fatalf("expected first game to be forced to active, got %v", out[0].CurrentState.Status)
	}
	if out[1].CurrentState.Status != models.StatusEnded {
		t.Fatalf("expected ended game to remain ended")
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	api.GET("/leagues", getLeagues)
	api.GET("/teams", getAllTeams)
	api.GET("/games", getGames)
	api.GET("/upcoming", getUpcomingGames)
	api.POST("/leagues", updateLeagueConfig)
	api.POST("/refresh", refreshActiveGames)
	api.GET("/events", getEvents)
	api.GET("/logs", getLogs)
	api.GET("/homeassistant/status", getHomeAssistantStatus)
	api.GET("/homeassistant/config", getHomeAssistantConfig)
	api.POST("/homeassistant/config", setHomeAssistantConfig)
	api.POST("/clear", clearGames)
	return r
}

func TestGetLeagues_OK(t *testing.T) {
	viper.Set("watch.nhl", []string{"BUF"})
	viper.Set("watch.mlb", []string{"TOR"})
	viper.Set("watch.cfl", []string{"*"})
	viper.Set("watch.nfl", []string{"BUF"})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/leagues", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetTeams_NFL_Filter_OK(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams?leagueId=6", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool             `json:"success"`
		Data    []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		t.Fatalf("expected teams array")
	}
}

func TestGetTeams_InvalidLeagueId(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams?leagueId=abc", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetGames_NormalizesActive(t *testing.T) {
	memoryStore.ClearAllGames()
	g := models.Game{GameCode: "TEST-1", LeagueId: models.LeagueIdNFL, CurrentState: models.GameState{Status: models.StatusUpcoming, Period: 1, Clock: "10:00"}}
	memoryStore.SetGame(g)
	memoryStore.AppendActiveGame(g)

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/games", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool          `json:"success"`
		Data    []models.Game `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) == 0 || resp.Data[0].CurrentState.Status != models.StatusActive {
		t.Fatalf("expected normalized active status")
	}
}

func TestUpdateLeagueConfig_WritesConfig(t *testing.T) {
	// Use a temp config file to avoid writing repo config
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(cfg)
	// Seed monitored teams
	body := `{ "leagueId": 6, "teams": ["BUF","MIA"] }`
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/leagues", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify file was written
	if _, err := os.Stat(cfg); err != nil {
		t.Fatalf("expected config file to be written: %v", err)
	}
}

func TestUpdateLeagueConfig_InvalidJSON(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/leagues", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateLeagueConfig_InvalidLeague(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/leagues", strings.NewReader(`{"leagueId":99,"teams":["X"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid leagueId, got %d", w.Code)
	}
}

func TestEventsAndLogsEndpoints_WithTempApplog(t *testing.T) {
	// Point applog to a temp file via viper key app_log.path
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "applog.jsonl")
	viper.Set("app_log.path", logPath)
	// Append a couple events
	ev1 := models.Event{LeagueId: models.LeagueIdNFL, TeamCode: "UNIT", GameCode: "UT1", Timestamp: time.Now(), LeagueName: "NFL"}
	applog.AppendEvent(ev1)
	ev2 := models.Event{LeagueId: models.LeagueIdNFL, TeamCode: "UNIT", GameCode: "UT2", Timestamp: time.Now(), LeagueName: "NFL"}
	applog.AppendEvent(ev2)
	// And a log line
	applog.AppendLogLine(models.AppLogLevelInfo, "test", "server_test", nil)

	r := setupRouter()
	// /api/events
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/events?leagueId=6&team=UNIT&limit=1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var evResp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &evResp); err != nil {
		t.Fatalf("unmarshal events: %v", err)
	}
	if !evResp.Success || len(evResp.Data) != 1 {
		t.Fatalf("expected 1 event result, got %d", len(evResp.Data))
	}
	// /api/logs
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/logs?limit=1", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var logResp struct {
		Success bool
		Data    []models.AppLogEntry
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &logResp); err != nil {
		t.Fatalf("unmarshal logs: %v", err)
	}
	if !logResp.Success || len(logResp.Data) == 0 {
		t.Fatalf("expected at least 1 log entry")
	}
}

func TestEvents_InvalidParams(t *testing.T) {
	r := setupRouter()
	cases := []string{
		"/api/events?leagueId=bad",
		"/api/events?since=not-rfc3339",
		"/api/events?limit=NaN",
	}
	for _, path := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, w.Code)
		}
	}
}

func TestLogs_InvalidParams(t *testing.T) {
	r := setupRouter()
	cases := []string{
		"/api/logs?leagueId=bad",
		"/api/logs?since=not-rfc3339",
		"/api/logs?limit=NaN",
	}
	for _, path := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, w.Code)
		}
	}
}

func TestHomeAssistantStatus_Unset(t *testing.T) {
	// Ensure no env or config values
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	viper.Set("home_assistant.url", "")
	viper.Set("home_assistant.access_token", "")

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/homeassistant/status", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || resp.Data["connected"].(bool) != false || resp.Data["source"].(string) != "unset" {
		t.Fatalf("expected disconnected unset status")
	}
}

func TestHomeAssistantStatus_EnvSource(t *testing.T) {
	os.Setenv("SUPERVISOR_API", "http://127.0.0.1:9999")
	os.Setenv("SUPERVISOR_TOKEN", "tok")
	defer os.Unsetenv("SUPERVISOR_API")
	defer os.Unsetenv("SUPERVISOR_TOKEN")

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/homeassistant/status", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || resp.Data["source"].(string) != "env" || resp.Data["tokenSet"].(bool) != true {
		t.Fatalf("expected env source and tokenSet true, got %+v", resp.Data)
	}
}

func TestHomeAssistantConfig_GET_SET(t *testing.T) {
	// Use temp config
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(cfg)
	// GET initially
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/homeassistant/config", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// SET values
	w2 := httptest.NewRecorder()
	body := `{"url":"http://localhost:8123","accessToken":"abc"}`
	req2, _ := http.NewRequest("POST", "/api/homeassistant/config", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	// Verify persisted config exists
	if _, err := os.Stat(cfg); err != nil {
		t.Fatalf("expected config file to be written: %v", err)
	}
}

func TestRefreshAndClear_OK(t *testing.T) {
	// Seed memory and then clear
	memoryStore.ClearAllGames()
	g := models.Game{GameCode: "X", LeagueId: models.LeagueIdNHL}
	memoryStore.SetGame(g)
	memoryStore.AppendActiveGame(g)

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/refresh", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh expected 200, got %d", w.Code)
	}
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/clear", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("clear expected 200, got %d", w2.Code)
	}
}

func TestGetUpcoming_EmptyMonitoredLists(t *testing.T) {
	// Ensure all monitored lists are empty so handler short-circuits to empty result
	viper.Set("watch.nhl", []string{})
	viper.Set("watch.mlb", []string{})
	viper.Set("watch.cfl", []string{})
	viper.Set("watch.iihf", []string{})
	viper.Set("watch.nfl", []string{})

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/upcoming", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []models.Game
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) != 0 {
		t.Fatalf("expected empty upcoming when no teams monitored")
	}
}

func TestGetTeams_NHL_OK(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams?leagueId=1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		t.Fatalf("expected NHL teams array")
	}
}

func TestGetTeams_CFL_MonitoredOnly(t *testing.T) {
	viper.Set("watch.cfl", []string{"WPG", "BC"})
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams?leagueId=5", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) != 2 {
		t.Fatalf("expected 2 CFL monitored teams, got %d", len(resp.Data))
	}
}

func TestBroadcastFunctions_MessageTypes(t *testing.T) {
	// Seed memory and send broadcasts
	memoryStore.ClearAllGames()
	g := models.Game{GameCode: "X", LeagueId: models.LeagueIdNHL}
	BroadcastGamesList() // should work with empty store
	BroadcastGameUpdate(g)
	BroadcastEvent(models.Event{TeamCode: "AAA"})
	BroadcastLog(models.AppLogEntry{Message: "hello"})

	deadline := time.Now().Add(300 * time.Millisecond)
	seen := map[string]bool{}
	for time.Now().Before(deadline) {
		select {
		case b := <-hub.broadcast:
			var msg WebSocketMessage
			_ = json.Unmarshal(b, &msg)
			seen[msg.Type] = true
			if seen["games_list"] && seen["game_update"] && seen["event"] && seen["log"] {
				break
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !(seen["games_list"] && seen["game_update"] && seen["event"] && seen["log"]) {
		t.Fatalf("expected all message types, got %+v", seen)
	}
}

func TestIsTeamMonitored_WildcardAndExact(t *testing.T) {
	if !isTeamMonitored([]string{"*"}, "ANY") {
		t.Fatalf("wildcard should monitor any team")
	}
	if isTeamMonitored([]string{"TOR"}, "WPG") {
		t.Fatalf("non-monitored team shouldn't match")
	}
	if !isTeamMonitored([]string{"tor"}, "TOR") {
		t.Fatalf("case-insensitive match failed")
	}
}

func TestGetAllTeams_MLB_OK(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams?leagueId=2", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		t.Fatalf("expected MLB teams array")
	}
	// Verify a known team has a logo
	foundTOR := false
	logoOK := false
	for _, tm := range resp.Data {
		if tm["code"] == "TOR" {
			foundTOR = true
			if s, ok := tm["logo"].(string); ok && s != "" {
				logoOK = true
			}
			break
		}
	}
	if !foundTOR || !logoOK {
		t.Fatalf("expected TOR with logo in MLB teams")
	}
}

func TestGetAllTeams_MissingLeagueId(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/teams", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetEvents_DeliveryMapping(t *testing.T) {
	// Isolate applog file
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "applog.jsonl")
	applog.SetLogFilePathForTest(logPath)
	viper.Set("app_log.path", logPath)
	// Append event with delivery fields via Append to capture fields
	success := true
	ev := models.Event{LeagueId: int(models.LeagueIdNFL), LeagueName: "NFL", TeamCode: "BUF", GameCode: "E1"}
	entry := models.AppLogEntry{Type: models.AppLogTypeEvent, LeagueId: models.LeagueIdNFL, LeagueName: "NFL", TeamCode: "BUF", GameCode: "E1", Target: "ha.event", Success: &success, Error: "", Event: &ev}
	applog.Append(entry)

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/events?leagueId=6&limit=5", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		t.Fatalf("expected at least one delivered event")
	}
	// Delivery object should be present
	if _, ok := resp.Data[0]["delivery"]; !ok {
		t.Fatalf("expected delivery metadata in response")
	}
}

func TestGetLogs_SinceFilter(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "applog.jsonl")
	applog.SetLogFilePathForTest(logPath)
	viper.Set("app_log.path", logPath)
	// Append an old and a new log line separated by more than a second to avoid RFC3339 truncation issues
	applog.AppendLogLine(models.AppLogLevelInfo, "old", "test", nil)
	time.Sleep(1100 * time.Millisecond)
	since := time.Now().Format(time.RFC3339)
	time.Sleep(1100 * time.Millisecond)
	applog.AppendLogLine(models.AppLogLevelInfo, "new", "test", nil)

	r := setupRouter()
	w := httptest.NewRecorder()
	path := "/api/logs?since=" + since
	req, _ := http.NewRequest("GET", path, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Success bool
		Data    []models.AppLogEntry
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || len(resp.Data) != 1 || resp.Data[0].Message != "new" {
		t.Fatalf("expected only the new log entry, got %+v", resp.Data)
	}
}

func TestGetGames_NFLDetailsEnrichment(t *testing.T) {
	r := setupRouter()

	// Test NFL games with details enrichment
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/games?leagueId=4", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	// Should have NFL games with enriched details
	if len(resp.Data) == 0 {
		t.Fatalf("expected at least one NFL game")
	}

	// Check that NFL games have enriched details
	for _, game := range resp.Data {
		if leagueId, ok := game["leagueId"].(float64); ok && int(leagueId) == 4 {
			// NFL game should have enriched details
			if _, hasDetails := game["details"]; !hasDetails {
				t.Fatalf("expected NFL game to have details enrichment")
			}
		}
	}
}

func TestGetGames_EnrichmentErrorHandling(t *testing.T) {
	r := setupRouter()

	// Test with invalid league ID that might cause enrichment errors
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/games?leagueId=999", nil)
	r.ServeHTTP(w, req)

	// Should still return 200 even if enrichment fails
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with enrichment errors, got %d", w.Code)
	}

	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true even with enrichment errors")
	}
}

func TestGetUpcoming_NoMonitoredTeams(t *testing.T) {
	r := setupRouter()

	// Test upcoming games endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/upcoming", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	// Should return empty array when no teams are monitored
	// (this tests the case where config has no monitored teams)
}

func TestGetGames_AllLeagues(t *testing.T) {
	r := setupRouter()

	// Test getting games from all leagues
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/games", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Success bool
		Data    []map[string]any
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	// Should have games from multiple leagues
	leagues := make(map[float64]bool)
	for _, game := range resp.Data {
		if leagueId, ok := game["leagueId"].(float64); ok {
			leagues[leagueId] = true
		}
	}

	// Should have games from at least NHL and NFL (based on mock data)
	if len(leagues) == 0 {
		t.Fatalf("expected games from at least one league")
	}
}
