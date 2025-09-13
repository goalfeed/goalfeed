package applog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/targets/notify"
	"goalfeed/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var logger = utils.GetLogger()

var (
	logFilePath string
	fileOnce    sync.Once
	fileMu      sync.Mutex
)

func getLogFilePath() string {
	if logFilePath != "" {
		return logFilePath
	}
	fileOnce.Do(func() {
		// Default path relative to working directory
		defaultPath := "app.log.jsonl"
		if p := config.GetString("app_log.path"); p != "" {
			defaultPath = p
		}
		// Ensure directory exists
		dir := filepath.Dir(defaultPath)
		if dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		logFilePath = defaultPath
	})
	return logFilePath
}

// Append writes an AppLogEntry to the durable log and broadcasts it to clients
func Append(entry models.AppLogEntry) {
	entry.Timestamp = time.Now()
	// Ensure ID
	if entry.Id == "" {
		entry.Id = fmt.Sprintf("%d-%s-%s", entry.Timestamp.UnixNano(), entry.TeamCode, entry.Metric)
	}

	// Serialize
	b, err := json.Marshal(entry)
	if err != nil {
		logger.Warn(fmt.Sprintf("applog marshal failed: %v", err))
		return
	}

	// Append to file
	fileMu.Lock()
	defer fileMu.Unlock()
	f, err := os.OpenFile(getLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logger.Warn(fmt.Sprintf("applog open failed: %v", err))
		return
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		logger.Warn(fmt.Sprintf("applog write failed: %v", err))
		return
	}

	// Broadcast via notify if available
	if notify.BroadcastLog != nil {
		notify.BroadcastLog(entry)
	}
}

// AppendLogLine appends a generic log line (debug/info/warn/error)
func AppendLogLine(level models.AppLogLevel, message, source string, fields map[string]string) {
	e := models.AppLogEntry{
		Type:    models.AppLogTypeLogLine,
		Level:   level,
		Source:  source,
		Message: message,
	}
	if len(fields) > 0 {
		var parts []string
		for k, v := range fields {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		e.Message = fmt.Sprintf("%s (%s)", e.Message, strings.Join(parts, ", "))
	}
	Append(e)
}

// Query returns recent log entries with optional filtering
func Query(leagueId int, teamCode string, since time.Time, limit int) []models.AppLogEntry {
	path := getLogFilePath()
	f, err := os.Open(path)
	if err != nil {
		return []models.AppLogEntry{}
	}
	defer f.Close()

	var results []models.AppLogEntry
	scanner := bufio.NewScanner(f)
	// Increase buffer for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var e models.AppLogEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if leagueId > 0 && int(e.LeagueId) != leagueId {
			continue
		}
		if teamCode != "" && !strings.EqualFold(e.TeamCode, teamCode) {
			continue
		}
		if !since.IsZero() && e.Timestamp.Before(since) {
			continue
		}
		results = append(results, e)
	}
	// If limit specified and smaller than results, slice the tail
	if limit > 0 && len(results) > limit {
		results = results[len(results)-limit:]
	}
	return results
}

// AppendEvent is a helper to log a domain event
func AppendEvent(ev models.Event) {
	entry := models.AppLogEntry{
		Type:       models.AppLogTypeEvent,
		LeagueId:   models.League(ev.LeagueId),
		LeagueName: ev.LeagueName,
		TeamCode:   ev.TeamCode,
		Opponent:   ev.OpponentCode,
		GameCode:   ev.GameCode,
		Event:      &ev,
	}
	Append(entry)
}

// AppendStateChange logs a team metric change with before/after values
func AppendStateChange(leagueId models.League, leagueName, teamCode, opponent, gameCode, metric string, before, after interface{}) {
	entry := models.AppLogEntry{
		Type:       models.AppLogTypeStateChange,
		LeagueId:   leagueId,
		LeagueName: leagueName,
		TeamCode:   teamCode,
		Opponent:   opponent,
		GameCode:   gameCode,
		Metric:     metric,
		Before:     before,
		After:      after,
	}
	Append(entry)
}
