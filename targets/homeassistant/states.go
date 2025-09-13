package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/targets/applog"
	"goalfeed/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var stateLogger = utils.GetLogger()

type entityCacheEntry struct {
	Serialized string
	UpdatedAt  time.Time
}

var (
	cacheMu       sync.Mutex
	entityCache   = map[string]entityCacheEntry{}
	debounceAfter = 500 * time.Millisecond
)

func getHAAuth() (string, string) {
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
	return homeAssistantURL, accessToken
}

func toStateString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "on"
		}
		return "off"
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case time.Time:
		if v.IsZero() {
			return "unknown"
		}
		return v.Format(time.RFC3339)
	default:
		if v == nil {
			return "unknown"
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func publishEntity(domain, entity, friendly string, state interface{}, attrs map[string]interface{}) (bool, string) {
	// Always include friendly_name
	if attrs == nil {
		attrs = map[string]interface{}{}
	}
	if friendly != "" {
		attrs["friendly_name"] = friendly
	}

	payload := map[string]interface{}{
		"state":      toStateString(state),
		"attributes": attrs,
	}
	serialized, _ := json.Marshal(payload)

	key := domain + "." + entity
	// Dedupe + debounce
	cacheMu.Lock()
	last, ok := entityCache[key]
	if ok {
		if string(serialized) == last.Serialized {
			cacheMu.Unlock()
			return false, ""
		}
		if time.Since(last.UpdatedAt) < debounceAfter {
			// IMPORTANT: Persist the new state in cache even if we debounce the outbound HA call.
			// This prevents repeated unknown->X transitions in our logs on subsequent publishes.
			entityCache[key] = entityCacheEntry{Serialized: string(serialized), UpdatedAt: time.Now()}
			cacheMu.Unlock()
			return false, ""
		}
	}
	// Update local cache so we persist state even if HA is unavailable
	entityCache[key] = entityCacheEntry{Serialized: string(serialized), UpdatedAt: time.Now()}
	cacheMu.Unlock()

	// Resolve HA endpoint after caching
	haURL, token := getHAAuth()
	if haURL == "" || token == "" {
		stateLogger.Warn("Home Assistant not configured; cached state only")
		return false, ""
	}

	url := fmt.Sprintf("%s/api/states/%s.%s", haURL, domain, entity)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(serialized))
	if err != nil {
		stateLogger.Warn(fmt.Sprintf("HA state req err: %v", err))
		return false, ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		stateLogger.Warn(fmt.Sprintf("HA state send err: %v", err))
		return false, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		stateLogger.Warn(fmt.Sprintf("HA state non-2xx: %s %s", resp.Status, string(bodyBytes)))
		return false, ""
	}
	// Extract previous state
	prev := ""
	if ok && last.Serialized != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(last.Serialized), &m); err == nil {
			if s, ok2 := m["state"].(string); ok2 {
				prev = s
			}
		}
	}
	return true, prev
}

func leagueSlug(league models.League) string {
	switch league {
	case models.LeagueIdNHL:
		return "nhl"
	case models.LeagueIdMLB:
		return "mlb"
	case models.LeagueIdCFL:
		return "cfl"
	case models.LeagueIdNFL:
		return "nfl"
	default:
		return "misc"
	}
}

func buildEntityName(league models.League, teamCode, metric string) string {
	base := fmt.Sprintf("goalfeed_%s_%s_%s", leagueSlug(league), strings.ToLower(teamCode), metric)
	return sanitizeId(base)
}

func PublishTeamSensors(game models.Game) {
	// Common metrics for both teams
	publishTeamCommon(game, game.CurrentState.Home, game.CurrentState.Away, game)
	publishTeamCommon(game, game.CurrentState.Away, game.CurrentState.Home, game)

	// League-specific metrics
	switch game.LeagueId {
	case models.LeagueIdMLB:
		publishMLBTeam(game, game.CurrentState.Home, game.CurrentState.Away)
		publishMLBTeam(game, game.CurrentState.Away, game.CurrentState.Home)
	case models.LeagueIdNFL, models.LeagueIdCFL:
		publishFootballTeam(game, game.CurrentState.Home, game.CurrentState.Away)
		publishFootballTeam(game, game.CurrentState.Away, game.CurrentState.Home)
	case models.LeagueIdNHL:
		publishNHLTeam(game, game.CurrentState.Home, game.CurrentState.Away)
		publishNHLTeam(game, game.CurrentState.Away, game.CurrentState.Home)
	}
}

func statusString(s models.GameStatus) string {
	switch s {
	case models.StatusUpcoming:
		return "scheduled"
	case models.StatusActive:
		return "active"
	case models.StatusDelayed:
		return "delayed"
	case models.StatusEnded:
		return "final"
	default:
		return "unknown"
	}
}

func publishTeamCommon(game models.Game, team models.TeamState, opponent models.TeamState, full models.Game) {
	league := game.LeagueId
	teamCode := team.Team.TeamCode
	oppCode := opponent.Team.TeamCode

	// team.status
	publishSensor(league, teamCode, "team.status", statusString(game.CurrentState.Status), map[string]interface{}{
		"opponent": oppCode,
	})

	// team.has_game_today
	hasGameToday := false
	if !game.GameDetails.GameDate.IsZero() {
		now := time.Now().In(game.GameDetails.GameDate.Location())
		hasGameToday = game.GameDetails.GameDate.Year() == now.Year() && game.GameDetails.GameDate.YearDay() == now.YearDay()
	} else {
		// Active games imply today
		hasGameToday = game.CurrentState.Status == models.StatusActive || game.CurrentState.Status == models.StatusUpcoming
	}
	publishBinarySensor(league, teamCode, "team.has_game_today", hasGameToday, map[string]interface{}{})

	// team.has_active_game
	hasActive := game.CurrentState.Status == models.StatusActive || (game.CurrentState.Period > 0 || (game.CurrentState.Clock != "" && game.CurrentState.Clock != "TBD"))
	publishBinarySensor(league, teamCode, "team.has_active_game", hasActive, nil)

	// team.current_score
	publishSensor(league, teamCode, "team.current_score", team.Score, nil)

	// team.opponent
	publishSensor(league, teamCode, "team.opponent", oppCode, nil)

	// team.home_away
	homeAway := "away"
	if team.Team.TeamCode == game.CurrentState.Home.Team.TeamCode {
		homeAway = "home"
	}
	publishSensor(league, teamCode, "team.home_away", homeAway, nil)

	// team.next_game_date (when known)
	if !game.GameDetails.GameDate.IsZero() {
		publishSensor(league, teamCode, "team.next_game_date", game.GameDetails.GameDate, nil)
	}

	// Clock/period
	if game.CurrentState.Clock != "" {
		publishSensor(league, teamCode, "team.clock", game.CurrentState.Clock, nil)
	}
	if game.CurrentState.Period > 0 {
		// Use sport-agnostic name; HA users can alias
		publishSensor(league, teamCode, "team.period", game.CurrentState.Period, nil)
	}
}

func publishMLBTeam(game models.Game, team models.TeamState, opponent models.TeamState) {
	league := game.LeagueId
	teamCode := team.Team.TeamCode
	d := game.CurrentState.Details

	// is_batting heuristic: if Possession equals team code, or Batter present for team
	isBatting := d.Possession == teamCode
	publishBinarySensor(league, teamCode, "team.is_batting", isBatting, nil)

	if d.BallCount > 0 || d.StrikeCount > 0 || d.Outs > 0 {
		publishSensor(league, teamCode, "team.balls", d.BallCount, nil)
		publishSensor(league, teamCode, "team.strikes", d.StrikeCount, nil)
		publishSensor(league, teamCode, "team.outs", d.Outs, nil)
	}

	if d.Bases != "" {
		publishSensor(league, teamCode, "team.runners_on_base", d.Bases, nil)
	}

	if d.Pitcher.Name != "" && !isBatting {
		publishSensor(league, teamCode, "team.current_pitcher", d.Pitcher.Name, nil)
	}
	if d.Batter.Name != "" && isBatting {
		publishSensor(league, teamCode, "team.current_batter", d.Batter.Name, nil)
	}
}

func publishFootballTeam(game models.Game, team models.TeamState, opponent models.TeamState) {
	league := game.LeagueId
	teamCode := team.Team.TeamCode
	d := game.CurrentState.Details

	hasPossession := strings.EqualFold(d.Possession, teamCode)
	publishBinarySensor(league, teamCode, "team.has_possession", hasPossession, nil)

	if d.Down > 0 {
		publishSensor(league, teamCode, "team.down", d.Down, nil)
	}
	if d.Distance > 0 {
		publishSensor(league, teamCode, "team.distance", d.Distance, nil)
	}
	if d.YardLine > 0 {
		publishSensor(league, teamCode, "team.yard_line", d.YardLine, nil)
	}

	// Heuristic red zone: offense and at or inside 20
	redZone := hasPossession && d.YardLine >= 80 // assuming 0-100 scale; if 1-50, this might be wrong
	publishBinarySensor(league, teamCode, "team.red_zone", redZone, nil)
}

func publishNHLTeam(game models.Game, team models.TeamState, opponent models.TeamState) {
	league := game.LeagueId
	teamCode := team.Team.TeamCode

	if team.Statistics.Shots > 0 {
		publishSensor(league, teamCode, "team.shots", team.Statistics.Shots, nil)
	}
	if team.Statistics.Penalties >= 0 {
		publishSensor(league, teamCode, "team.penalties", team.Statistics.Penalties, nil)
	}
	// power_play unknown without explicit flag; skip to avoid flapping
}

func publishSensor(league models.League, teamCode, metric string, value interface{}, attrs map[string]interface{}) {
	entity := buildEntityName(league, teamCode, metric)
	friendly := fmt.Sprintf("%s %s", teamCode, strings.ReplaceAll(metric, "_", " "))
	// log before HA publish
	key := "sensor." + entity
	cacheMu.Lock()
	last, ok := entityCache[key]
	cacheMu.Unlock()
	prev := ""
	if ok && last.Serialized != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(last.Serialized), &m); err == nil {
			if s, ok2 := m["state"].(string); ok2 {
				prev = s
			}
		}
	}
	if prev != toStateString(value) {
		applog.AppendStateChange(league, strings.ToUpper(leagueSlug(league)), teamCode, "", "", metric, prev, value)
	}
	_, _ = publishEntity("sensor", entity, friendly, value, attrs)
}

func publishBinarySensor(league models.League, teamCode, metric string, value bool, attrs map[string]interface{}) {
	entity := buildEntityName(league, teamCode, metric)
	friendly := fmt.Sprintf("%s %s", teamCode, strings.ReplaceAll(metric, "_", " "))
	key := "binary_sensor." + entity
	cacheMu.Lock()
	last, ok := entityCache[key]
	cacheMu.Unlock()
	prev := ""
	if ok && last.Serialized != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(last.Serialized), &m); err == nil {
			if s, ok2 := m["state"].(string); ok2 {
				prev = s
			}
		}
	}
	if prev != toStateString(value) {
		applog.AppendStateChange(league, strings.ToUpper(leagueSlug(league)), teamCode, "", "", metric, prev, value)
	}
	_, _ = publishEntity("binary_sensor", entity, friendly, value, attrs)
}

func PublishGameSummary(game models.Game) {
	// Disabled to avoid per-game entity sprawl. Left for potential on-demand summaries.
}

// PublishScheduleSensorsForGame publishes schedule-related sensors for both teams for an upcoming game
func PublishScheduleSensorsForGame(game models.Game) {
	// Team status should be scheduled for upcoming
	// Home
	publishTeamSchedule(game.LeagueId, game.CurrentState.Home, game.CurrentState.Away, game)
	// Away
	publishTeamSchedule(game.LeagueId, game.CurrentState.Away, game.CurrentState.Home, game)
}

func publishTeamSchedule(league models.League, team models.TeamState, opponent models.TeamState, game models.Game) {
	teamCode := team.Team.TeamCode
	oppCode := opponent.Team.TeamCode
	// status -> scheduled
	publishSensor(league, teamCode, "team.status", "scheduled", map[string]interface{}{
		"opponent": oppCode,
	})
	// has_active_game -> false
	publishBinarySensor(league, teamCode, "team.has_active_game", false, nil)
	// has_game_today
	hasGameToday := false
	if !game.GameDetails.GameDate.IsZero() {
		now := time.Now().In(game.GameDetails.GameDate.Location())
		hasGameToday = game.GameDetails.GameDate.Year() == now.Year() && game.GameDetails.GameDate.YearDay() == now.YearDay()
	}
	publishBinarySensor(league, teamCode, "team.has_game_today", hasGameToday, nil)
	// next_game_date
	if !game.GameDetails.GameDate.IsZero() {
		publishSensor(league, teamCode, "team.next_game_date", game.GameDetails.GameDate, nil)
	}
	// opponent
	publishSensor(league, teamCode, "team.opponent", oppCode, nil)
	// home_away
	homeAway := "away"
	if team.Team.TeamCode == game.CurrentState.Home.Team.TeamCode {
		homeAway = "home"
	}
	publishSensor(league, teamCode, "team.home_away", homeAway, nil)
}

func sanitizeId(s string) string {
	s = strings.ToLower(s)
	var b []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b = append(b, r)
		} else {
			b = append(b, '_')
		}
	}
	return string(b)
}

func PublishBaselineForMonitoredTeams() {
	leagues := []struct {
		id  models.League
		key string
	}{
		{models.LeagueIdNHL, "nhl"},
		{models.LeagueIdMLB, "mlb"},
		{models.LeagueIdCFL, "cfl"},
		{models.LeagueIdNFL, "nfl"},
	}
	for _, lc := range leagues {
		teams := config.GetStringSlice("watch." + lc.key)
		if len(teams) == 0 {
			continue
		}
		for _, t := range teams {
			if t == "*" || strings.TrimSpace(t) == "" {
				continue
			}
			// Common baseline entities
			publishSensor(lc.id, t, "team.status", "idle", nil)
			publishBinarySensor(lc.id, t, "team.has_game_today", false, nil)
			publishBinarySensor(lc.id, t, "team.has_active_game", false, nil)
			publishSensor(lc.id, t, "team.current_score", 0, nil)
			publishSensor(lc.id, t, "team.opponent", "unknown", nil)
			publishSensor(lc.id, t, "team.home_away", "unknown", nil)
			publishSensor(lc.id, t, "team.next_game_date", "unknown", nil)
			publishSensor(lc.id, t, "team.clock", "unknown", nil)
			publishSensor(lc.id, t, "team.period", 0, nil)

			// League-specific baseline entities
			switch lc.id {
			case models.LeagueIdMLB:
				publishBinarySensor(lc.id, t, "team.is_batting", false, nil)
				publishSensor(lc.id, t, "team.balls", 0, nil)
				publishSensor(lc.id, t, "team.strikes", 0, nil)
				publishSensor(lc.id, t, "team.outs", 0, nil)
				publishSensor(lc.id, t, "team.runners_on_base", "", nil)
				publishSensor(lc.id, t, "team.current_pitcher", "unknown", nil)
				publishSensor(lc.id, t, "team.current_batter", "unknown", nil)
			case models.LeagueIdNFL, models.LeagueIdCFL:
				publishBinarySensor(lc.id, t, "team.has_possession", false, nil)
				publishSensor(lc.id, t, "team.down", 0, nil)
				publishSensor(lc.id, t, "team.distance", 0, nil)
				publishSensor(lc.id, t, "team.yard_line", 0, nil)
				publishBinarySensor(lc.id, t, "team.red_zone", false, nil)
			case models.LeagueIdNHL:
				publishSensor(lc.id, t, "team.shots", 0, nil)
				publishSensor(lc.id, t, "team.penalties", 0, nil)
			}
		}
	}
}

// PublishEndOfGameReset resets dynamic sensors when a game ends, and marks status final
func PublishEndOfGameReset(game models.Game) {
	// Helper to decide has_game_today based on game date
	isToday := func(t time.Time) bool {
		if t.IsZero() {
			return false
		}
		now := time.Now().In(t.Location())
		return t.Year() == now.Year() && t.YearDay() == now.YearDay()
	}

	resetTeam := func(league models.League, teamCode string) {
		publishSensor(league, teamCode, "team.status", "final", nil)
		publishBinarySensor(league, teamCode, "team.has_active_game", false, nil)
		publishBinarySensor(league, teamCode, "team.has_game_today", isToday(game.GameDetails.GameDate), nil)
		publishSensor(league, teamCode, "team.clock", "", nil)
		publishSensor(league, teamCode, "team.period", 0, nil)
		// League-specific resets
		switch league {
		case models.LeagueIdMLB:
			publishBinarySensor(league, teamCode, "team.is_batting", false, nil)
			publishSensor(league, teamCode, "team.balls", 0, nil)
			publishSensor(league, teamCode, "team.strikes", 0, nil)
			publishSensor(league, teamCode, "team.outs", 0, nil)
			publishSensor(league, teamCode, "team.runners_on_base", "", nil)
			publishSensor(league, teamCode, "team.current_pitcher", "unknown", nil)
			publishSensor(league, teamCode, "team.current_batter", "unknown", nil)
		case models.LeagueIdNFL, models.LeagueIdCFL:
			publishBinarySensor(league, teamCode, "team.has_possession", false, nil)
			publishSensor(league, teamCode, "team.down", 0, nil)
			publishSensor(league, teamCode, "team.distance", 0, nil)
			publishSensor(league, teamCode, "team.yard_line", 0, nil)
			publishBinarySensor(league, teamCode, "team.red_zone", false, nil)
		case models.LeagueIdNHL:
			// Keep shots/penalties as final numbers; nothing to reset
		}
	}

	resetTeam(game.LeagueId, game.CurrentState.Home.Team.TeamCode)
	resetTeam(game.LeagueId, game.CurrentState.Away.Team.TeamCode)
}
