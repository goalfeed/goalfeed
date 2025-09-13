package applog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goalfeed/models"
)

// override log file path for tests via package-level variable
func TestAppendAndQueryUsingTempFile(t *testing.T) {
	// Use temp dir
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log.jsonl")
	// set global once value by directly assigning computed path
	logFilePath = path

	// Append different types
	team := "UNITTEST"
	game := "g-unit"
	AppendLogLine(models.AppLogLevelInfo, "hello", "test", map[string]string{"k": "v"})
	AppendStateChange(models.LeagueIdNHL, "NHL", team, "TOR", game, "score", 1, 2)
	AppendEvent(models.Event{Id: "e1", LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL", TeamCode: team, OpponentCode: "TOR", GameCode: game})

	// Query by our team to avoid entries from other tests
	resTeam := Query(int(models.LeagueIdNHL), team, time.Time{}, 0)
	if len(resTeam) != 2 { // state_change + event
		t.Fatalf("expected 2 entries for %s, got %d", team, len(resTeam))
	}

	// Limit applied to team-filtered results
	resLimited := Query(int(models.LeagueIdNHL), team, time.Time{}, 1)
	if len(resLimited) != 1 {
		t.Fatalf("expected 1 limited entry, got %d", len(resLimited))
	}

	// Ensure file exists and is non-empty
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	if st.Size() == 0 {
		t.Fatalf("expected log file to be non-empty")
	}
}

func TestQuerySinceAndLeagueOnly(t *testing.T) {
	// Use temp file and clear global path
	dir := t.TempDir()
	path := filepath.Join(dir, "since.log.jsonl")
	logFilePath = path

	// Controlled timestamps
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now().Add(-1 * time.Minute)

	// Prepare explicit entries
	entries := []models.AppLogEntry{
		{Id: "old-nhl", LeagueId: models.LeagueIdNHL, LeagueName: "NHL", TeamCode: "ABC", GameCode: "1", Timestamp: oldTime},
		{Id: "new-nhl", LeagueId: models.LeagueIdNHL, LeagueName: "NHL", TeamCode: "ABC", GameCode: "2", Timestamp: newTime},
		{Id: "new-mlb", LeagueId: models.LeagueIdMLB, LeagueName: "MLB", TeamCode: "DEF", GameCode: "3", Timestamp: newTime},
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	for _, e := range entries {
		b, _ := json.Marshal(e)
		_, _ = f.Write(append(b, '\n'))
	}
	_ = f.Close()

	// Since filter should include only the two new entries when league unconstrained
	since := time.Now().Add(-10 * time.Minute)
	res := Query(0, "", since, 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 entries since cutoff, got %d", len(res))
	}
	// League-only filter for NHL should return only the NHL new one
	resNHL := Query(int(models.LeagueIdNHL), "", since, 0)
	if len(resNHL) != 1 {
		t.Fatalf("expected 1 NHL entry since cutoff, got %d", len(resNHL))
	}
}
