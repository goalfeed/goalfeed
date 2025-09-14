package nfl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"goalfeed/models"
	"goalfeed/targets/memoryStore"

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
	// Use the refactored version for better testability
	applyNFLPatchesRefactored(pl, topic)
}
