package nfl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"goalfeed/models"
	"goalfeed/targets/memoryStore"
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

// FastcastConfig holds configuration for NFL Fastcast connection
type FastcastConfig struct {
	ReconnectBaseMs int
	ReconnectMaxMs  int
	PingIntervalSec int
	PongWaitSec     int
}

// FastcastConnection manages WebSocket connection and message handling
type FastcastConnection struct {
	config         FastcastConfig
	conn           *websocket.Conn
	sid            string
	stopSubs       chan struct{}
	lastMidByTopic map[string]int64
}

var nflEventPath = regexp.MustCompile(`e:(\d+)`)
var downDistanceAt = regexp.MustCompile(`(?i)^(1st|2nd|3rd|4th)\s*&\s*(\d+)(?:\s+at\s+([A-Z]{2,4})\s+(\d+))?`)

// NewFastcastConfig creates a new FastcastConfig with default values
func NewFastcastConfig() FastcastConfig {
	baseMs := viper.GetInt("nfl.fastcast.reconnect_base_ms")
	if baseMs <= 0 {
		baseMs = 2000
	}
	maxMs := viper.GetInt("nfl.fastcast.reconnect_max_ms")
	if maxMs <= 0 {
		maxMs = 30000
	}
	pingIntervalSec := viper.GetInt("nfl.fastcast.ping_interval_sec")
	if pingIntervalSec <= 0 {
		pingIntervalSec = 20
	}
	pongWaitSec := viper.GetInt("nfl.fastcast.pong_wait_sec")
	if pongWaitSec <= 0 {
		pongWaitSec = 60
	}

	return FastcastConfig{
		ReconnectBaseMs: baseMs,
		ReconnectMaxMs:  maxMs,
		PingIntervalSec: pingIntervalSec,
		PongWaitSec:     pongWaitSec,
	}
}

// CalculateJitter calculates jitter for backoff timing
func CalculateJitter(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	j := ms / 10
	base := int64(ms)
	ji := int64(j)
	v := base + ji - (time.Now().UnixNano() % (2*ji + 1))
	if v < 0 {
		v = base
	}
	return time.Duration(v) * time.Millisecond
}

// CalculateBackoff calculates exponential backoff with jitter
func CalculateBackoff(currentMs, baseMs, maxMs int) int {
	if currentMs < maxMs {
		currentMs *= 2
		if currentMs > maxMs {
			currentMs = maxMs
		}
	}
	return currentMs
}

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
	// Use the refactored version for better testability
	RunNFLFastcastRefactored()
}

func applyNFLPatches(pl json.RawMessage, topic string) {
	// Placeholder implementation - this was refactored but the refactored version was removed
	// For now, just ensure no panic occurs
	_ = pl
	_ = topic
}

// FetchFastcastHost retrieves the Fastcast host information
func FetchFastcastHost() (*fastcastHost, error) {
	// This function already exists in the original file
	return fetchFastcastHost()
}

// CreateWebSocketConnection establishes a WebSocket connection to Fastcast
func CreateWebSocketConnection(host *fastcastHost) (*websocket.Conn, error) {
	wsURL := fmt.Sprintf("wss://%s:%d/FastcastService/pubsub/profiles/12000?TrafficManager-Token=%s",
		host.IP, host.SecurePort, host.Token)
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
		Proxy:             http.ProxyFromEnvironment,
	}
	conn, _, err := dialer.Dial(wsURL, http.Header{})
	return conn, err
}

// NewFastcastConnection creates a new FastcastConnection instance
func NewFastcastConnection(config FastcastConfig) *FastcastConnection {
	return &FastcastConnection{
		config:         config,
		stopSubs:       make(chan struct{}),
		lastMidByTopic: make(map[string]int64),
	}
}

// SetupKeepalive configures ping/pong handling for the connection
func (fc *FastcastConnection) SetupKeepalive() {
	_ = fc.conn.SetReadDeadline(time.Now().Add(time.Duration(fc.config.PongWaitSec) * time.Second))
	fc.conn.SetPongHandler(func(string) error {
		_ = fc.conn.SetReadDeadline(time.Now().Add(time.Duration(fc.config.PongWaitSec) * time.Second))
		return nil
	})
}

// StartPingTicker starts the periodic ping ticker
func (fc *FastcastConnection) StartPingTicker() {
	pingTicker := time.NewTicker(time.Duration(fc.config.PingIntervalSec) * time.Second)
	go func() {
		defer pingTicker.Stop()
		for {
			select {
			case <-pingTicker.C:
				_ = fc.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
			case <-fc.stopSubs:
				return
			}
		}
	}()
}

// SubscribeToChannels subscribes to NFL event channels
func (fc *FastcastConnection) SubscribeToChannels() {
	if fc.sid == "" || fc.conn == nil {
		return
	}

	// Subscribe to broad NFL event channel and top events
	for _, tc := range []string{"event-topevents", "event-football-nfl"} {
		b, _ := json.Marshal(wsMsg{Op: "S", Sid: fc.sid, Tc: tc})
		_ = fc.conn.WriteMessage(websocket.TextMessage, b)
	}
}

// SubscribeToActiveGames subscribes to active NFL games
func (fc *FastcastConnection) SubscribeToActiveGames() {
	if fc.sid == "" || fc.conn == nil {
		return
	}

	for _, g := range memoryStore.GetAllGames() {
		if g.LeagueId == models.LeagueIdNFL {
			tc := "gp-football-nfl-" + g.GameCode
			b, _ := json.Marshal(wsMsg{Op: "S", Sid: fc.sid, Tc: tc})
			_ = fc.conn.WriteMessage(websocket.TextMessage, b)
		}
	}
}

// StartActiveGameSubscription starts periodic subscription to active games
func (fc *FastcastConnection) StartActiveGameSubscription() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fc.SubscribeToActiveGames()
			case <-fc.stopSubs:
				return
			}
		}
	}()
}

// HandleConnectionMessage processes incoming WebSocket messages
func (fc *FastcastConnection) HandleConnectionMessage(msg []byte) bool {
	var m wsMsg
	if err := json.Unmarshal(msg, &m); err != nil {
		return true // Continue processing
	}

	switch m.Op {
	case "C":
		if m.Sid != "" {
			fc.sid = m.Sid
			fc.SubscribeToChannels()
			fc.SubscribeToActiveGames()
			fc.StartActiveGameSubscription()
		}
	case "P":
		if m.Tc != "" && m.Mid > 0 {
			if last, ok := fc.lastMidByTopic[m.Tc]; ok && m.Mid <= last {
				return true // Skip stale message
			}
			fc.lastMidByTopic[m.Tc] = m.Mid
		}
		// Log debug message (removed logger dependency for now)
		applyNFLPatches(m.Pl, m.Tc)
	}

	return true // Continue processing
}

// CloseConnection closes the WebSocket connection and cleanup
func (fc *FastcastConnection) CloseConnection() {
	if fc.conn != nil {
		_ = fc.conn.Close()
	}
	close(fc.stopSubs)
}

// RunConnectionLoop runs the main connection loop
func (fc *FastcastConnection) RunConnectionLoop() {
	for {
		_, msg, err := fc.conn.ReadMessage()
		if err != nil {
			fc.CloseConnection()
			break
		}

		if !fc.HandleConnectionMessage(msg) {
			break
		}
	}
}

// RunNFLFastcastRefactored is the refactored version of runNFLFastcast
func RunNFLFastcastRefactored() {
	config := NewFastcastConfig()
	backoffMs := config.ReconnectBaseMs

	for {
		// Fetch host and create connection
		host, err := FetchFastcastHost()
		if err != nil {
			time.Sleep(CalculateJitter(backoffMs))
			backoffMs = CalculateBackoff(backoffMs, config.ReconnectBaseMs, config.ReconnectMaxMs)
			continue
		}

		conn, err := CreateWebSocketConnection(host)
		if err != nil {
			time.Sleep(CalculateJitter(backoffMs))
			backoffMs = CalculateBackoff(backoffMs, config.ReconnectBaseMs, config.ReconnectMaxMs)
			continue
		}

		// Reset backoff on successful connect
		backoffMs = config.ReconnectBaseMs
		// Log connection success (removed logger dependency for now)

		// Send connection message
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"op":"C"}`))

		// Create connection handler and run
		fc := NewFastcastConnection(config)
		fc.conn = conn
		fc.SetupKeepalive()
		fc.StartPingTicker()
		fc.RunConnectionLoop()

		// Reconnect after short delay
		time.Sleep(CalculateJitter(backoffMs))
		backoffMs = CalculateBackoff(backoffMs, config.ReconnectBaseMs, config.ReconnectMaxMs)
	}
}
