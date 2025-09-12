package applog

import (
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
	AppendLogLine(models.AppLogLevelInfo, "hello", "test", map[string]string{"k": "v"})
	AppendStateChange(models.LeagueIdNHL, "NHL", "WPG", "TOR", "g1", "score", 1, 2)
	AppendEvent(models.Event{Id: "e1", LeagueId: int(models.LeagueIdNHL), LeagueName: "NHL", TeamCode: "WPG", OpponentCode: "TOR", GameCode: "g1"})

	// Query all
	res := Query(0, "", time.Time{}, 0)
	if len(res) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(res))
	}

	// Filter by league
	res2 := Query(int(models.LeagueIdNHL), "", time.Time{}, 0)
	if len(res2) != 2 { // state_change + event (log line has no league)
		t.Fatalf("expected 2 NHL entries, got %d", len(res2))
	}

	// Filter by team
	res3 := Query(int(models.LeagueIdNHL), "WPG", time.Time{}, 0)
	if len(res3) != 2 {
		t.Fatalf("expected 2 entries for WPG, got %d", len(res3))
	}

	// Limit
	res4 := Query(0, "", time.Time{}, 2)
	if len(res4) != 2 {
		t.Fatalf("expected 2 limited entries, got %d", len(res4))
	}

	// Ensure file exists and is non-empty
	st, err := os.Stat(path)
	if err != nil || st.Size() == 0 {
		t.Fatalf("expected log file to exist and be non-empty: %v size=%d", err, st.Size())
	}
}
