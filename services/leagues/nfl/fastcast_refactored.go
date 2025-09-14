package nfl

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	nflClients "goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"goalfeed/targets/memoryStore"
	"goalfeed/targets/notify"
)

// wrapperFlex represents a flexible wrapper for Fastcast payloads
type wrapperFlex struct {
	Ts int64           `json:"ts"`
	C  int64           `json:"~c"`
	Pl json.RawMessage `json:"pl"`
}

// competitorMapping holds the mapping of competitor indices to sides and team IDs
type competitorMapping struct {
	Side   map[string]string // idx -> "home"|"away"
	TeamID map[string]string // idx -> teamId
}

// normalizePayload converts various payload formats to normalized bytes
func normalizePayload(pl json.RawMessage) []byte {
	var asString string
	if err := json.Unmarshal(pl, &asString); err == nil {
		return []byte(asString)
	}
	return pl
}

// decodeWrapperPayload extracts operations from a wrapper payload
func decodeWrapperPayload(normalized []byte) ([]patchOp, error) {
	var w wrapperFlex
	if err := json.Unmarshal(normalized, &w); err != nil || len(w.Pl) == 0 {
		return nil, fmt.Errorf("invalid wrapper")
	}

	// Try to decode as base64+zlib first
	var maybeB64 string
	if err := json.Unmarshal(w.Pl, &maybeB64); err == nil && maybeB64 != "" {
		return decodeBase64ZlibPayload(maybeB64)
	}

	// Try to decode as direct JSON array
	var ops []patchOp
	if err := json.Unmarshal(w.Pl, &ops); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wrapper pl")
	}
	return ops, nil
}

// decodeBase64ZlibPayload decodes a base64+zlib encoded payload
func decodeBase64ZlibPayload(b64 string) ([]patchOp, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	zr, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("zlib reader failed: %v", err)
	}
	defer zr.Close()

	inflated, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("zlib inflate failed: %v", err)
	}

	var ops []patchOp
	if err := json.Unmarshal(inflated, &ops); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %v", err)
	}
	return ops, nil
}

// decodeDirectPayload attempts to decode payload without wrapper
func decodeDirectPayload(normalized []byte) ([]patchOp, error) {
	// Try base64+zlib directly
	if decoded, err := base64.StdEncoding.DecodeString(string(normalized)); err == nil {
		if ops, err := decodeBase64ZlibPayload(string(decoded)); err == nil {
			return ops, nil
		}
	}

	// Try interpreting as []patchOp JSON
	var ops []patchOp
	if err := json.Unmarshal(normalized, &ops); err != nil {
		return nil, fmt.Errorf("failed to unmarshal as patch ops")
	}
	return ops, nil
}

// extractOperations extracts patch operations from various payload formats
func extractOperations(pl json.RawMessage) ([]patchOp, error) {
	normalized := normalizePayload(pl)

	// Try wrapper format first
	if ops, err := decodeWrapperPayload(normalized); err == nil {
		return ops, nil
	}

	// Try direct format
	return decodeDirectPayload(normalized)
}

// buildCompetitorMapping creates a mapping of competitor indices to sides and team IDs
func buildCompetitorMapping(ops []patchOp) competitorMapping {
	compSide := make(map[string]string)
	compTeam := make(map[string]string)

	reHomeAway := regexp.MustCompile(`/competitors/(\d+)/homeAway$`)
	reTeamId := regexp.MustCompile(`/competitors/(\d+)/team/id$`)

	for _, op := range ops {
		if m := reHomeAway.FindStringSubmatch(op.Path); len(m) == 2 {
			if v, ok := op.Value.(string); ok {
				compSide[m[1]] = strings.ToLower(v)
			}
		}
		if m := reTeamId.FindStringSubmatch(op.Path); len(m) == 2 {
			if v, ok := op.Value.(string); ok {
				compTeam[m[1]] = v
			}
		}
	}

	return competitorMapping{
		Side:   compSide,
		TeamID: compTeam,
	}
}

// extractEventID extracts the event ID from the operation path or topic
func extractEventID(op patchOp, topic string) string {
	m := nflEventPath.FindStringSubmatch(op.Path)
	if len(m) == 2 {
		return m[1]
	}

	// Fallback: extract from topic like gp-football-nfl-<eventId>
	if strings.HasPrefix(topic, "gp-football-nfl-") {
		return strings.TrimPrefix(topic, "gp-football-nfl-")
	}

	return ""
}

// findGameByEventID finds a game by its event ID
func findGameByEventID(eventID string) (models.Game, error) {
	for _, g := range memoryStore.GetAllGames() {
		if g.LeagueId == models.LeagueIdNFL && g.GameCode == eventID {
			gameKey := g.GetGameKey()
			return memoryStore.GetGameByGameKey(gameKey)
		}
	}
	return models.Game{}, fmt.Errorf("game not found for event ID: %s", eventID)
}

// applyClockUpdate applies clock-related updates to the game
func applyClockUpdate(game models.Game, op patchOp) models.Game {
	// Display clock updates
	if strings.HasSuffix(op.Path, "/fullStatus/displayClock") || strings.HasSuffix(op.Path, "/clock") {
		if v, ok := op.Value.(string); ok && v != "" {
			game.CurrentState.Clock = v
			game.CurrentState.Status = models.StatusActive
		}
	}

	// Numeric clock updates (seconds to M:SS)
	if strings.HasSuffix(op.Path, "/fullStatus/clock") {
		if vv, ok := op.Value.(float64); ok {
			sec := int(vv)
			m := sec / 60
			s := sec % 60
			game.CurrentState.Clock = fmt.Sprintf("%d:%02d", m, s)
			game.CurrentState.Status = models.StatusActive
		}
	}

	return game
}

// applyPeriodUpdate applies period-related updates to the game
func applyPeriodUpdate(game models.Game, op patchOp) models.Game {
	// Short detail updates (e.g., "6:14 - 3rd")
	if strings.HasSuffix(op.Path, "/fullStatus/type/shortDetail") {
		if v, ok := op.Value.(string); ok && v != "" {
			parts := strings.Split(v, " - ")
			if len(parts) == 2 {
				game.CurrentState.Clock = parts[0]
				q := strings.Fields(parts[1])
				if len(q) > 0 {
					period, periodType := parsePeriodFromText(q[0])
					if period > 0 {
						game.CurrentState.Period = period
						game.CurrentState.PeriodType = periodType
					}
				}
			}
		}
	}

	// Detail updates (e.g., "3:28 - 3rd Quarter")
	if strings.HasSuffix(op.Path, "/fullStatus/type/detail") || strings.HasSuffix(op.Path, "/summary") {
		if v, ok := op.Value.(string); ok && v != "" {
			parts := strings.Split(v, " - ")
			if len(parts) >= 2 {
				game.CurrentState.Clock = parts[0]
				q := strings.Fields(parts[1])
				if len(q) > 0 {
					period, periodType := parsePeriodFromText(q[0])
					if period > 0 {
						game.CurrentState.Period = period
						game.CurrentState.PeriodType = periodType
					}
				}
			}
		}
	}

	return game
}

// parsePeriodFromText parses period information from text
func parsePeriodFromText(text string) (int, string) {
	switch strings.ToLower(text) {
	case "1st":
		return 1, "QUARTER"
	case "2nd":
		return 2, "QUARTER"
	case "3rd":
		return 3, "QUARTER"
	case "4th":
		return 4, "QUARTER"
	case "ot", "ot1":
		return 5, "OVERTIME"
	case "2ot":
		return 6, "OVERTIME"
	default:
		return 0, ""
	}
}

// applySituationUpdate applies situation-related updates to the game
func applySituationUpdate(game models.Game, op patchOp) models.Game {
	// Down updates
	if strings.HasSuffix(op.Path, "/situation/down") {
		if vv, ok := op.Value.(float64); ok {
			d := int(vv)
			if d >= 1 && d <= 4 {
				game.CurrentState.Details.Down = d
			} else {
				game.CurrentState.Details.Down = 0
			}
		}
	}

	// Yard line updates
	if strings.HasSuffix(op.Path, "/situation/yardLine") {
		if vv, ok := op.Value.(float64); ok {
			game.CurrentState.Details.YardLine = int(vv)
		}
	}

	// Possession text updates
	if strings.HasSuffix(op.Path, "/possessionText") {
		if v, ok := op.Value.(string); ok && v != "" {
			parts := strings.Fields(v)
			if len(parts) >= 1 {
				game.CurrentState.Details.Possession = strings.ToUpper(parts[0])
			}
			if len(parts) >= 2 {
				if yl, err := strconv.Atoi(parts[1]); err == nil {
					game.CurrentState.Details.YardLine = yl
				}
			}
		}
	}

	// Down distance text updates
	if strings.HasSuffix(op.Path, "/situation/shortDownDistanceText") || strings.HasSuffix(op.Path, "/downDistanceText") {
		if v, ok := op.Value.(string); ok && v != "" {
			if m2 := downDistanceAt.FindStringSubmatch(v); len(m2) >= 3 {
				order := strings.ToLower(m2[1])
				down := 0
				switch order {
				case "1st":
					down = 1
				case "2nd":
					down = 2
				case "3rd":
					down = 3
				case "4th":
					down = 4
				}
				if down >= 1 && down <= 4 {
					game.CurrentState.Details.Down = down
				} else {
					game.CurrentState.Details.Down = 0
				}
				if dist, err := strconv.Atoi(m2[2]); err == nil {
					game.CurrentState.Details.Distance = dist
				}
				if len(m2) >= 5 {
					if m2[3] != "" {
						game.CurrentState.Details.Possession = strings.ToUpper(m2[3])
					}
					if yl, err := strconv.Atoi(m2[4]); err == nil && yl > 0 {
						game.CurrentState.Details.YardLine = yl
					}
				}
			}
		}
	}

	return game
}

// applyScoreUpdate applies score updates to the game
func applyScoreUpdate(game models.Game, op patchOp, mapping competitorMapping) models.Game {
	reScore := regexp.MustCompile(`/competitors/(\d+)/score$`)
	msc := reScore.FindStringSubmatch(op.Path)
	if len(msc) != 2 {
		return game
	}

	idx := msc[1]
	sv, ok := op.Value.(string)
	if !ok {
		return game
	}

	n, err := strconv.Atoi(sv)
	if err != nil {
		return game
	}

	side := mapping.Side[idx]
	teamID := mapping.TeamID[idx]

	// Try team ID mapping first
	if teamID != "" {
		if game.CurrentState.Home.Team.ExtID == teamID {
			game.CurrentState.Home.Score = n
			return game
		} else if game.CurrentState.Away.Team.ExtID == teamID {
			game.CurrentState.Away.Score = n
			return game
		}
	}

	// Fall back to side mapping
	if side == "home" {
		game.CurrentState.Home.Score = n
		return game
	} else if side == "away" {
		game.CurrentState.Away.Score = n
		return game
	}

	// Unknown side - fall back to fresh data from API
	return refreshScoreFromAPI(game, op.Path)
}

// refreshScoreFromAPI refreshes the score from the NFL API
func refreshScoreFromAPI(game models.Game, path string) models.Game {
	// Only refresh on summary/lastPlay as a safety net
	if !strings.HasSuffix(path, "/summary") && !strings.Contains(path, "/situation/lastPlay") {
		return game
	}

	svc := NFLService{Client: nflClients.NFLAPIClient{}}
	fresh := svc.GameFromScoreboard(game.GameCode)
	if fresh.GameCode != "" {
		game.CurrentState.Home.Score = fresh.CurrentState.Home.Score
		game.CurrentState.Away.Score = fresh.CurrentState.Away.Score
	}
	return game
}

// broadcastGameUpdate broadcasts the game update if needed
func broadcastGameUpdate(topic string) {
	if notify.BroadcastGame == nil {
		return
	}

	if !strings.HasPrefix(topic, "gp-football-nfl-") {
		return
	}

	eventID := strings.TrimPrefix(topic, "gp-football-nfl-")
	for _, g := range memoryStore.GetAllGames() {
		if g.LeagueId == models.LeagueIdNFL && g.GameCode == eventID {
			notify.BroadcastGame(g)
			break
		}
	}
}

// applyNFLPatchesRefactored is the refactored version of applyNFLPatches
func applyNFLPatchesRefactored(pl json.RawMessage, topic string) {
	// Extract operations from payload
	ops, err := extractOperations(pl)
	if err != nil {
		logger.Warnf("NFL Fastcast: failed to extract operations: %v", err)
		return
	}

	// Build competitor mapping
	mapping := buildCompetitorMapping(ops)

	// Process each operation
	for _, op := range ops {
		// Extract event ID
		eventID := extractEventID(op, topic)
		if eventID == "" {
			logger.Warnf("NFL Fastcast: unable to determine eventID from topic=%s path=%s", topic, op.Path)
			continue
		}

		// Find the game
		game, err := findGameByEventID(eventID)
		if err != nil {
			logger.Debugf("NFL Fastcast: event %s not in active memory store; skipping", eventID)
			continue
		}

		// Apply updates based on path
		game = applyClockUpdate(game, op)
		game = applyPeriodUpdate(game, op)
		game = applySituationUpdate(game, op)
		game = applyScoreUpdate(game, op, mapping)

		// Save the game
		memoryStore.SetGame(game)
	}

	// Broadcast the update
	broadcastGameUpdate(topic)
}
