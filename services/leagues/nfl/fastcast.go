package nfl

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	nflClients "goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"goalfeed/targets/memoryStore"
	"goalfeed/targets/notify"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type fastcastHost struct {
	IP         string `json:"ip"`
	SecurePort int    `json:"securePort"`
	Token      string `json:"token"`
}

type wsMsg struct {
	Op  string          `json:"op"`
	Sid string          `json:"sid,omitempty"`
	Tc  string          `json:"tc,omitempty"`
	Pl  json.RawMessage `json:"pl,omitempty"`
	Mid int64           `json:"mid,omitempty"`
}

type patchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

var nflEventPath = regexp.MustCompile(`e:(\d+)`)
var downDistanceAt = regexp.MustCompile(`(?i)^(1st|2nd|3rd|4th)\s*&\s*(\d+)(?:\s+at\s+([A-Z]{2,4})\s+(\d+))?`)

func fetchFastcastHost() (*fastcastHost, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://fastcast.semfs.engsvc.go.com/public/websockethost", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Goalfeed)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fastcast bootstrap status %s", resp.Status)
	}
	var h fastcastHost
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, err
	}
	return &h, nil
}

// StartNFLFastcast starts a background listener that updates NFL games using ESPN Fastcast
func StartNFLFastcast() {
	if !viper.GetBool("nfl.fastcast.enabled") {
		return
	}
	go runNFLFastcast()
}

func runNFLFastcast() {
	// Reconnect backoff parameters
	baseMs := viper.GetInt("nfl.fastcast.reconnect_base_ms")
	if baseMs <= 0 {
		baseMs = 2000
	}
	maxMs := viper.GetInt("nfl.fastcast.reconnect_max_ms")
	if maxMs <= 0 {
		maxMs = 30000
	}
	backoffMs := baseMs
	jitter := func(ms int) time.Duration {
		j := ms / 10
		base := int64(ms)
		ji := int64(j)
		v := base + ji - (time.Now().UnixNano() % (2*ji + 1))
		if v < 0 {
			v = base
		}
		return time.Duration(v) * time.Millisecond
	}
	for {
		h, err := fetchFastcastHost()
		if err != nil {
			time.Sleep(jitter(backoffMs))
			if backoffMs < maxMs {
				backoffMs *= 2
				if backoffMs > maxMs {
					backoffMs = maxMs
				}
			}
			continue
		}
		wsURL := fmt.Sprintf("wss://%s:%d/FastcastService/pubsub/profiles/12000?TrafficManager-Token=%s", h.IP, h.SecurePort, h.Token)
		dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, EnableCompression: true, Proxy: http.ProxyFromEnvironment}
		conn, _, err := dialer.Dial(wsURL, http.Header{})
		if err != nil {
			time.Sleep(jitter(backoffMs))
			if backoffMs < maxMs {
				backoffMs *= 2
				if backoffMs > maxMs {
					backoffMs = maxMs
				}
			}
			continue
		}
		// reset backoff on successful connect
		backoffMs = baseMs
		logger.Infof("NFL Fastcast connected: %s", wsURL)
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"op":"C"}`))
		// read loop
		var sid string
		stopSubs := make(chan struct{})
		// Keepalive: set read deadlines and start periodic ping
		pingIntervalSec := viper.GetInt("nfl.fastcast.ping_interval_sec")
		if pingIntervalSec <= 0 {
			pingIntervalSec = 20
		}
		pongWaitSec := viper.GetInt("nfl.fastcast.pong_wait_sec")
		if pongWaitSec <= 0 {
			pongWaitSec = 60
		}
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(pongWaitSec) * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(pongWaitSec) * time.Second))
			return nil
		})
		pingTicker := time.NewTicker(time.Duration(pingIntervalSec) * time.Second)
		go func() {
			defer pingTicker.Stop()
			for {
				select {
				case <-pingTicker.C:
					_ = conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
				case <-stopSubs:
					return
				}
			}
		}()
		// Track last message id per topic to drop stale/out-of-order batches
		lastMidByTopic := map[string]int64{}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				_ = conn.Close()
				close(stopSubs)
				break
			}
			var m wsMsg
			if err := json.Unmarshal(msg, &m); err != nil {
				continue
			}
			switch m.Op {
			case "C":
				if m.Sid != "" {
					sid = m.Sid
					// Subscribe to broad NFL event channel and top events
					for _, tc := range []string{"event-topevents", "event-football-nfl"} {
						b, _ := json.Marshal(wsMsg{Op: "S", Sid: sid, Tc: tc})
						_ = conn.WriteMessage(websocket.TextMessage, b)
					}
					// Subscribe to active NFL games now and periodically
					subscribeActiveNFLGames := func() {
						for _, g := range memoryStore.GetAllGames() {
							if g.LeagueId == models.LeagueIdNFL {
								tc := "gp-football-nfl-" + g.GameCode
								b, _ := json.Marshal(wsMsg{Op: "S", Sid: sid, Tc: tc})
								_ = conn.WriteMessage(websocket.TextMessage, b)
							}
						}
					}
					subscribeActiveNFLGames()
					go func() {
						ticker := time.NewTicker(15 * time.Second)
						defer ticker.Stop()
						for {
							select {
							case <-ticker.C:
								subscribeActiveNFLGames()
							case <-stopSubs:
								return
							}
						}
					}()
				}
			case "P":
				if m.Tc != "" && m.Mid > 0 {
					if last, ok := lastMidByTopic[m.Tc]; ok && m.Mid <= last {
						continue
					}
					lastMidByTopic[m.Tc] = m.Mid
				}
				logger.Debugf("NFL Fastcast: received patch topic=%s mid=%d bytes=%d", m.Tc, m.Mid, len(m.Pl))
				applyNFLPatches(m.Pl, m.Tc)
			}
		}
		// reconnect after short delay
		time.Sleep(jitter(backoffMs))
		if backoffMs < maxMs {
			backoffMs *= 2
			if backoffMs > maxMs {
				backoffMs = maxMs
			}
		}
	}
}

func applyNFLPatches(pl json.RawMessage, topic string) {
	// Robustly decode Fastcast payloads. Supported cases:
	// 1) pl is a JSON-encoded string of an object: {"ts":..., "~c":0|1, "pl":[ops] | "<base64+zlib>"}
	// 2) pl is that object directly (not quoted)
	// 3) pl is already a base64+zlib string of ops
	// 4) pl is already a JSON array of ops

	// Normalize to bytes of a potential wrapper or ops
	var normalized []byte
	var asString string
	if err := json.Unmarshal(pl, &asString); err == nil {
		normalized = []byte(asString)
	} else {
		normalized = pl
	}

	// Try wrapper with flexible pl type
	type wrapperFlex struct {
		Ts int64           `json:"ts"`
		C  int64           `json:"~c"`
		Pl json.RawMessage `json:"pl"`
	}

	var ops []patchOp
	var w wrapperFlex
	if err := json.Unmarshal(normalized, &w); err == nil && len(w.Pl) > 0 {
		// If w.Pl is a JSON string (base64+zlib) decode it; else if it's an array, unmarshal directly
		var maybeB64 string
		if err := json.Unmarshal(w.Pl, &maybeB64); err == nil && maybeB64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(maybeB64)
			if err != nil {
				return
			}
			zr, err := zlib.NewReader(bytes.NewReader(decoded))
			if err != nil {
				return
			}
			inflated, err := io.ReadAll(zr)
			_ = zr.Close()
			if err != nil {
				return
			}
			if err := json.Unmarshal(inflated, &ops); err != nil {
				return
			}
		} else {
			if err := json.Unmarshal(w.Pl, &ops); err != nil {
				return
			}
		}
	} else {
		// No wrapper. Try base64+zlib directly
		if decoded, err := base64.StdEncoding.DecodeString(string(normalized)); err == nil {
			if zr, err2 := zlib.NewReader(bytes.NewReader(decoded)); err2 == nil {
				inflated, err3 := io.ReadAll(zr)
				_ = zr.Close()
				if err3 == nil && json.Unmarshal(inflated, &ops) == nil {
					// ok
				} else {
					return
				}
			} else {
				return
			}
		} else {
			// Finally, try interpreting normalized as []patchOp JSON
			if err := json.Unmarshal(normalized, &ops); err != nil {
				return
			}
		}
	}

	// Pre-scan to map competitor index to side/team id if present in this batch
	compSide := map[string]string{} // idx -> "home"|"away"
	compTeam := map[string]string{} // idx -> teamId
	reHomeAway := regexp.MustCompile(`/competitors/(\d+)/homeAway$`)
	reTeamId := regexp.MustCompile(`/competitors/(\d+)/team/id$`)
	reScore := regexp.MustCompile(`/competitors/(\d+)/score$`)
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
		_ = reScore // ensure used later
	}

	for _, op := range ops {
		m := nflEventPath.FindStringSubmatch(op.Path)
		var eventID string
		if len(m) == 2 {
			eventID = m[1]
		} else {
			// Fallback: extract from topic like gp-football-nfl-<eventId>
			if strings.HasPrefix(topic, "gp-football-nfl-") {
				eventID = strings.TrimPrefix(topic, "gp-football-nfl-")
			}
			if eventID == "" {
				// Unable to map patch to an event/game; log for visibility
				logger.Warnf("NFL Fastcast: unable to determine eventID from topic=%s path=%s", topic, op.Path)
				continue
			}
		}

		// Find the corresponding game
		var gameKey string
		for _, g := range memoryStore.GetAllGames() {
			if g.LeagueId == models.LeagueIdNFL && g.GameCode == eventID {
				gameKey = g.GetGameKey()
				break
			}
		}
		if gameKey == "" {
			logger.Debugf("NFL Fastcast: event %s not in active memory store; skipping", eventID)
			continue
		}
		game, err := memoryStore.GetGameByGameKey(gameKey)
		if err != nil {
			continue
		}

		updated := false
		// Map known paths
		if strings.HasSuffix(op.Path, "/fullStatus/displayClock") || strings.HasSuffix(op.Path, "/clock") {
			if v, ok := op.Value.(string); ok && v != "" {
				game.CurrentState.Clock = v
				game.CurrentState.Status = models.StatusActive
				updated = true
			}
		}
		if strings.HasSuffix(op.Path, "/fullStatus/clock") {
			switch vv := op.Value.(type) {
			case float64:
				// Convert seconds to M:SS
				sec := int(vv)
				m := sec / 60
				s := sec % 60
				game.CurrentState.Clock = fmt.Sprintf("%d:%02d", m, s)
				game.CurrentState.Status = models.StatusActive
				updated = true
			}
		}
		if strings.HasSuffix(op.Path, "/fullStatus/type/shortDetail") {
			if v, ok := op.Value.(string); ok && v != "" {
				// Example: "6:14 - 3rd"
				parts := strings.Split(v, " - ")
				if len(parts) == 2 {
					game.CurrentState.Clock = parts[0]
					q := strings.Fields(parts[1])
					if len(q) > 0 {
						switch strings.ToLower(q[0]) {
						case "1st":
							game.CurrentState.Period = 1
							game.CurrentState.PeriodType = "QUARTER"
						case "2nd":
							game.CurrentState.Period = 2
							game.CurrentState.PeriodType = "QUARTER"
						case "3rd":
							game.CurrentState.Period = 3
							game.CurrentState.PeriodType = "QUARTER"
						case "4th":
							game.CurrentState.Period = 4
							game.CurrentState.PeriodType = "QUARTER"
						case "ot", "ot1":
							game.CurrentState.Period = 5
							game.CurrentState.PeriodType = "OVERTIME"
						case "2ot":
							game.CurrentState.Period = 6
							game.CurrentState.PeriodType = "OVERTIME"
						}
						updated = true
					}
				}
			}
		}
		if strings.HasSuffix(op.Path, "/fullStatus/type/detail") || strings.HasSuffix(op.Path, "/summary") {
			if v, ok := op.Value.(string); ok && v != "" {
				// Example: "3:28 - 3rd Quarter" or "6:14 - 3rd"
				parts := strings.Split(v, " - ")
				if len(parts) >= 2 {
					game.CurrentState.Clock = parts[0]
					q := strings.Fields(parts[1])
					if len(q) > 0 {
						val := strings.ToLower(q[0])
						switch val {
						case "1st":
							game.CurrentState.Period = 1
						case "2nd":
							game.CurrentState.Period = 2
						case "3rd":
							game.CurrentState.Period = 3
						case "4th":
							game.CurrentState.Period = 4
						case "ot", "ot1":
							game.CurrentState.Period = 5
						case "2ot":
							game.CurrentState.Period = 6
						}
						if game.CurrentState.Period > 0 {
							game.CurrentState.PeriodType = "QUARTER"
							if game.CurrentState.Period >= 5 {
								game.CurrentState.PeriodType = "OVERTIME"
							}
						}
						updated = true
					}
				}
			}
		}
		if strings.HasSuffix(op.Path, "/situation/down") {
			switch vv := op.Value.(type) {
			case float64:
				d := int(vv)
				if d >= 1 && d <= 4 {
					game.CurrentState.Details.Down = d
				} else {
					game.CurrentState.Details.Down = 0
				}
				updated = true
			}
		}
		if strings.HasSuffix(op.Path, "/situation/yardLine") {
			switch vv := op.Value.(type) {
			case float64:
				game.CurrentState.Details.YardLine = int(vv)
				updated = true
			}
		}
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
				updated = true
			}
		}
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
					updated = true
				}
			}
		}
		// Score updates: try to apply directly when competitors/<idx>/score present, preferring compTeam when available
		if msc := reScore.FindStringSubmatch(op.Path); len(msc) == 2 {
			idx := msc[1]
			if sv, ok := op.Value.(string); ok {
				if n, err := strconv.Atoi(sv); err == nil {
					side := compSide[idx]
					teamID := compTeam[idx]
					applied := false
					if teamID != "" {
						if game.CurrentState.Home.Team.ExtID == teamID {
							game.CurrentState.Home.Score = n
							updated = true
							applied = true
						} else if game.CurrentState.Away.Team.ExtID == teamID {
							game.CurrentState.Away.Score = n
							updated = true
							applied = true
						}
					}
					if !applied {
						if side == "home" {
							game.CurrentState.Home.Score = n
							updated = true
						} else if side == "away" {
							game.CurrentState.Away.Score = n
							updated = true
						} else {
							// Unknown side; fall back to summary to avoid mis-assigning
							svc := NFLService{Client: nflClients.NFLAPIClient{}}
							fresh := svc.GameFromScoreboard(eventID)
							if fresh.GameCode != "" {
								game.CurrentState.Home.Score = fresh.CurrentState.Home.Score
								game.CurrentState.Away.Score = fresh.CurrentState.Away.Score
								updated = true
							}
						}
					}
				}
			}
		}
		// Also refresh on summary/lastPlay as a safety net after scoring plays
		if strings.HasSuffix(op.Path, "/summary") || strings.Contains(op.Path, "/situation/lastPlay") {
			svc := NFLService{Client: nflClients.NFLAPIClient{}}
			fresh := svc.GameFromScoreboard(eventID)
			if fresh.GameCode != "" {
				game.CurrentState.Home.Score = fresh.CurrentState.Home.Score
				game.CurrentState.Away.Score = fresh.CurrentState.Away.Score
				updated = true
			}
		}

		if updated {
			memoryStore.SetGame(game)
			// Defer notification to after batch to avoid spamming
		}
	}
	// After applying the batch for this topic/game, broadcast once
	if notify.BroadcastGame != nil {
		if strings.HasPrefix(topic, "gp-football-nfl-") {
			eventID := strings.TrimPrefix(topic, "gp-football-nfl-")
			for _, g := range memoryStore.GetAllGames() {
				if g.LeagueId == models.LeagueIdNFL && g.GameCode == eventID {
					notify.BroadcastGame(g)
					break
				}
			}
		}
	}
}
