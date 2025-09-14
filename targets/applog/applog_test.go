package applog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"goalfeed/models"
	"goalfeed/targets/notify"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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

func TestQuery_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.log.jsonl")
	logFilePath = path

	// Create empty file
	_, err := os.Create(path)
	if err != nil {
		t.Fatalf("create empty file: %v", err)
	}

	// Query should return empty results
	res := Query(0, "", time.Time{}, 0)
	if len(res) != 0 {
		t.Fatalf("expected 0 entries from empty file, got %d", len(res))
	}
}

func TestQuery_NoMatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nomatch.log.jsonl")
	logFilePath = path

	// Write some test entries
	entries := []models.AppLogEntry{
		{
			Id:        "entry1",
			Timestamp: time.Now(),
			Level:     models.AppLogLevelInfo,
			Message:   "Test message 1",
			GameCode:  "game-1",
			LeagueId:  models.LeagueIdNHL,
			TeamCode:  "TEAM1",
		},
		{
			Id:        "entry2",
			Timestamp: time.Now(),
			Level:     models.AppLogLevelInfo,
			Message:   "Test message 2",
			GameCode:  "game-2",
			LeagueId:  models.LeagueIdNFL,
			TeamCode:  "TEAM2",
		},
	}

	// Write entries to file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}
		f.WriteString(string(data) + "\n")
	}
	f.Close()

	// Query for non-existent team should return empty
	res := Query(0, "NONEXISTENT", time.Time{}, 0)
	if len(res) != 0 {
		t.Fatalf("expected 0 entries for non-existent team, got %d", len(res))
	}

	// Query for non-existent league should return empty
	res = Query(999, "", time.Time{}, 0)
	if len(res) != 0 {
		t.Fatalf("expected 0 entries for non-existent league, got %d", len(res))
	}
}

func TestQuery_FilterCombinations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "filter-test.log.jsonl")
	logFilePath = path

	// Write test entries with different combinations
	entries := []models.AppLogEntry{
		{
			Id:        "old-entry",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Level:     models.AppLogLevelInfo,
			Message:   "Old message",
			GameCode:  "game-1",
			LeagueId:  models.LeagueIdNHL,
			TeamCode:  "TEAM1",
		},
		{
			Id:        "recent-error",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Level:     models.AppLogLevelError,
			Message:   "Recent error",
			GameCode:  "game-1",
			LeagueId:  models.LeagueIdNHL,
			TeamCode:  "TEAM1",
		},
		{
			Id:        "current-info",
			Timestamp: time.Now(),
			Level:     models.AppLogLevelInfo,
			Message:   "Current message",
			GameCode:  "game-2",
			LeagueId:  models.LeagueIdNFL,
			TeamCode:  "TEAM2",
		},
	}

	// Write entries to file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}
		f.WriteString(string(data) + "\n")
	}
	f.Close()

	// Test since filter
	since := time.Now().Add(-30 * time.Minute)
	res := Query(0, "", since, 0)
	if len(res) != 1 {
		t.Fatalf("expected 1 entry since cutoff, got %d", len(res))
	}

	// Test limit filter
	res = Query(0, "", time.Time{}, 2)
	if len(res) != 2 {
		t.Fatalf("expected 2 limited entries, got %d", len(res))
	}

	// Test team filter
	res = Query(0, "TEAM1", time.Time{}, 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 TEAM1 entries, got %d", len(res))
	}

	// Test league filter
	res = Query(int(models.LeagueIdNHL), "", time.Time{}, 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 NHL entries, got %d", len(res))
	}
}

func TestGetLogFilePath_DefaultPath(t *testing.T) {
	// Reset the global state
	logFilePath = ""
	fileOnce = sync.Once{}

	// Test default path
	path := getLogFilePath()
	assert.Equal(t, "app.log.jsonl", path)

	// Test that it's cached
	path2 := getLogFilePath()
	assert.Equal(t, path, path2)
}

func TestGetLogFilePath_WithConfig(t *testing.T) {
	// Reset the global state
	logFilePath = ""
	fileOnce = sync.Once{}

	// Set config path
	viper.Set("app_log.path", "/tmp/test.log")
	defer viper.Reset()

	path := getLogFilePath()
	assert.Equal(t, "/tmp/test.log", path)
}

func TestGetLogFilePath_WithCustomPath(t *testing.T) {
	// Reset the global state
	logFilePath = ""
	fileOnce = sync.Once{}

	// Set custom path directly
	logFilePath = "/custom/path.log"
	path := getLogFilePath()
	assert.Equal(t, "/custom/path.log", path)
}

func TestSetLogFilePathForTest(t *testing.T) {
	// Reset the global state
	logFilePath = ""

	// Test setting a custom path
	testPath := "/tmp/test-applog.log"
	SetLogFilePathForTest(testPath)

	// Verify it was set
	assert.Equal(t, testPath, logFilePath)

	// Test that getLogFilePath returns the set path
	path := getLogFilePath()
	assert.Equal(t, testPath, path)
}

func TestSetLogFilePathForTest_WithDirectory(t *testing.T) {
	// Reset the global state
	logFilePath = ""

	// Test setting a path with directory
	testPath := "/tmp/testdir/applog.log"
	SetLogFilePathForTest(testPath)

	// Verify it was set
	assert.Equal(t, testPath, logFilePath)

	// Verify directory was created
	dir := filepath.Dir(testPath)
	if dir != "." && dir != "" {
		_, err := os.Stat(dir)
		assert.NoError(t, err)
		// Clean up
		os.RemoveAll(dir)
	}
}

func TestAppend_WithExistingId(t *testing.T) {
	SetLogFilePathForTest(filepath.Join(t.TempDir(), "test.log"))
	defer os.Remove(getLogFilePath())

	entry := models.AppLogEntry{
		Id:       "custom-id",
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}
	Append(entry)

	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 1)
	assert.Equal(t, "custom-id", entries[0].Id)
}

func TestAppend_WithEmptyId(t *testing.T) {
	SetLogFilePathForTest(filepath.Join(t.TempDir(), "test.log"))
	defer os.Remove(getLogFilePath())

	entry := models.AppLogEntry{
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}
	Append(entry)

	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 1)
	assert.NotEmpty(t, entries[0].Id)
	assert.Contains(t, entries[0].Id, "TOR")
	assert.Contains(t, entries[0].Id, "goal")
}

func TestAppend_WithInvalidDirectory(t *testing.T) {
	// Set path to a directory that doesn't exist and can't be created
	SetLogFilePathForTest("/invalid/path/that/does/not/exist/test.log")

	entry := models.AppLogEntry{
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}

	// This should not panic, but may log warnings
	Append(entry)

	// Verify no entries were written due to file open error
	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 0)
}

func TestAppend_WithReadOnlyFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "readonly.log")

	// Create a read-only file
	f, err := os.Create(logFile)
	assert.NoError(t, err)
	f.Close()

	// Make it read-only
	err = os.Chmod(logFile, 0o444)
	assert.NoError(t, err)

	SetLogFilePathForTest(logFile)
	defer os.Remove(logFile)

	entry := models.AppLogEntry{
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}

	// This should not panic, but may log warnings due to write error
	Append(entry)

	// Verify no entries were written due to write error
	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 0)
}

func TestAppend_WithNotifyBroadcast(t *testing.T) {
	SetLogFilePathForTest(filepath.Join(t.TempDir(), "test.log"))
	defer os.Remove(getLogFilePath())

	// Test with notify.BroadcastLog set to nil (should not panic)
	originalBroadcast := notify.BroadcastLog
	notify.BroadcastLog = nil
	defer func() { notify.BroadcastLog = originalBroadcast }()

	entry := models.AppLogEntry{
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}

	// This should not panic even with nil BroadcastLog
	Append(entry)

	// Verify entry was written
	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 1)
}

func TestAppend_WithNotifyBroadcastFunction(t *testing.T) {
	SetLogFilePathForTest(filepath.Join(t.TempDir(), "test.log"))
	defer os.Remove(getLogFilePath())

	// Test with notify.BroadcastLog set to a function
	originalBroadcast := notify.BroadcastLog
	broadcastCalled := false
	var broadcastedEntry models.AppLogEntry

	notify.BroadcastLog = func(entry models.AppLogEntry) {
		broadcastCalled = true
		broadcastedEntry = entry
	}
	defer func() { notify.BroadcastLog = originalBroadcast }()

	entry := models.AppLogEntry{
		Type:     models.AppLogTypeEvent,
		LeagueId: models.LeagueIdNHL,
		TeamCode: "TOR",
		Metric:   "goal",
	}

	Append(entry)

	// Verify entry was written
	entries := Query(0, "", time.Time{}, 0)
	assert.Len(t, entries, 1)

	// Verify broadcast was called
	assert.True(t, broadcastCalled)
	assert.Equal(t, models.League(models.LeagueIdNHL), broadcastedEntry.LeagueId)
	assert.Equal(t, "TOR", broadcastedEntry.TeamCode)
}
