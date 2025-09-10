package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

func fetchFastcastHost() (*fastcastHost, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://fastcast.semfs.engsvc.go.com/public/websockethost", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Goalfeed fastcast-probe)")
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

// minimal structures for ESPN summary to get start time
type nflSummary struct {
	Events []struct {
		ID     string `json:"id"`
		Date   string `json:"date"`
		Status struct {
			Type struct {
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
		Competitions []struct {
			Date string `json:"date"`
		} `json:"competitions"`
	} `json:"events"`
}

func fetchStartTimeUTC(eventID string) (time.Time, error) {
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=%s", eventID)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Goalfeed fastcast-probe)")
	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()
	var sb nflSummary
	if err := json.NewDecoder(resp.Body).Decode(&sb); err != nil {
		return time.Time{}, err
	}
	if len(sb.Events) == 0 {
		return time.Time{}, fmt.Errorf("no events in summary for %s", eventID)
	}
	// prefer event date, fallback to competition date
	candidates := []string{sb.Events[0].Date}
	if len(sb.Events[0].Competitions) > 0 {
		candidates = append(candidates, sb.Events[0].Competitions[0].Date)
	}
	layouts := []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05-07:00", "2006-01-02T15:04Z", "2006-01-02T15:04-07:00"}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, c); err == nil {
				return t.UTC(), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("could not parse start time for %s", eventID)
}

// fallback to scoreboard to find event by id
func fetchStartTimeFromScoreboard(eventID string) (time.Time, error) {
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Goalfeed fastcast-probe)")
	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()
	var sb struct {
		Events []struct {
			ID           string `json:"id"`
			Date         string `json:"date"`
			Competitions []struct {
				Date string `json:"date"`
			} `json:"competitions"`
		} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sb); err != nil {
		return time.Time{}, err
	}
	layouts := []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05-07:00", "2006-01-02T15:04Z", "2006-01-02T15:04-07:00"}
	for _, ev := range sb.Events {
		if ev.ID == eventID {
			candidates := []string{ev.Date}
			if len(ev.Competitions) > 0 {
				candidates = append(candidates, ev.Competitions[0].Date)
			}
			for _, c := range candidates {
				if c == "" {
					continue
				}
				for _, layout := range layouts {
					if t, err := time.Parse(layout, c); err == nil {
						return t.UTC(), nil
					}
				}
			}
		}
	}
	return time.Time{}, fmt.Errorf("could not find event %s in scoreboard", eventID)
}

func fetchStartTimeFlexible(eventID string) (time.Time, error) {
	if t, err := fetchStartTimeUTC(eventID); err == nil && !t.IsZero() {
		return t, nil
	}
	return fetchStartTimeFromScoreboard(eventID)
}

func decodePatchPayload(pl json.RawMessage) ([]patchOp, error) {
	// pl can be either quoted JSON with { pl: base64 } or directly an object/string
	var raw string
	var wrapper struct {
		Ts int64  `json:"ts"`
		Pl string `json:"pl"`
	}
	if err := json.Unmarshal(pl, &wrapper); err == nil && wrapper.Pl != "" {
		raw = wrapper.Pl
	} else {
		raw = string(pl)
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	zr, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}
	inflated, err := io.ReadAll(zr)
	_ = zr.Close()
	if err != nil {
		return nil, err
	}
	var ops []patchOp
	if err := json.Unmarshal(inflated, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

func main() {
	var gameID string
	var durationStr string
	var outPath string
	var subscribeAll bool
	flag.StringVar(&gameID, "game", "", "ESPN NFL gameId (e.g., 401772810)")
	flag.StringVar(&durationStr, "duration", "1h", "Capture duration (e.g., 1h, 30m)")
	flag.StringVar(&outPath, "out", "", "Output log file path (default fastcast-<game>-<ts>.log.jsonl)")
	flag.BoolVar(&subscribeAll, "all", false, "Also subscribe to event-football-nfl and event-topevents")
	flag.Parse()

	if strings.TrimSpace(gameID) == "" {
		fmt.Fprintln(os.Stderr, "-game is required")
		os.Exit(2)
	}
	captureDur, err := time.ParseDuration(durationStr)
	if err != nil || captureDur <= 0 {
		fmt.Fprintln(os.Stderr, "invalid -duration; use forms like 1h or 45m")
		os.Exit(2)
	}

	startUTC, err := fetchStartTimeFlexible(gameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fetch start time: %v (will still capture and poll)\n", err)
	}

	now := time.Now().UTC()
	// Determine planned capture end time: one hour after start if known; else provisional now+duration
	endAt := now.Add(captureDur)
	if !startUTC.IsZero() {
		endAt = startUTC.Add(captureDur)
		// If game starts within ~2 hours, idle until 5 minutes before
		if startUTC.After(now) {
			untilStart := time.Until(startUTC)
			if untilStart > 2*time.Hour+5*time.Minute {
				fmt.Fprintf(os.Stderr, "game %s starts in %s (>2h); exiting (run closer to kickoff)\n", gameID, untilStart.Round(time.Minute))
				os.Exit(3)
			}
			wakeDelay := untilStart - 5*time.Minute
			if wakeDelay > 0 {
				fmt.Printf("Waiting %s until 5 minutes before kickoff (%s UTC)\n", wakeDelay.Round(time.Second), startUTC.Format(time.RFC3339))
				time.Sleep(wakeDelay)
			}
		} else {
			// already started; capture for requested duration from now
			endAt = now.Add(captureDur)
		}
	}

	if outPath == "" {
		ts := time.Now().Format("20060102-150405")
		outPath = filepath.Join(".", fmt.Sprintf("fastcast-%s-%s.log.jsonl", gameID, ts))
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	fmt.Printf("Logging Fastcast for game %s to %s for %s...\n", gameID, outPath, captureDur)

	// Establish Fastcast connection
	h, err := fetchFastcastHost()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch fastcast host: %v\n", err)
		os.Exit(1)
	}
	wsURL := fmt.Sprintf("wss://%s:%d/FastcastService/pubsub/profiles/12000?TrafficManager-Token=%s", h.IP, h.SecurePort, h.Token)
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, EnableCompression: true, Proxy: http.ProxyFromEnvironment}
	conn, _, err := dialer.Dial(wsURL, http.Header{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket dial failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("Connected to %s\n", wsURL)

	// Keepalive
	pongWait := 60 * time.Second
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { _ = conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	// Send connect op (log exactly what we send)
	connectMsg := []byte(`{"op":"C"}`)
	_, _ = outFile.Write(connectMsg)
	_, _ = outFile.Write([]byte("\n"))
	_ = conn.WriteMessage(websocket.TextMessage, connectMsg)

	done := make(chan struct{})

	// read loop
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// write the raw inbound message
			_, _ = outFile.Write(msg)
			_, _ = outFile.Write([]byte("\n"))

			// minimal parse to detect connect ack and subscribe
			var m wsMsg
			if err := json.Unmarshal(msg, &m); err == nil && m.Op == "C" && m.Sid != "" {
				// subscribe (and log sent subscription messages)
				topics := []string{fmt.Sprintf("gp-football-nfl-%s", gameID)}
				if subscribeAll {
					topics = append(topics, "event-football-nfl", "event-topevents")
				}
				for _, tc := range topics {
					b, _ := json.Marshal(wsMsg{Op: "S", Sid: m.Sid, Tc: tc})
					_, _ = outFile.Write(b)
					_, _ = outFile.Write([]byte("\n"))
					_ = conn.WriteMessage(websocket.TextMessage, b)
				}
			}
		}
	}()

	// If start time wasn't known, poll and extend endAt to (start+duration) once discovered
	if startUTC.IsZero() {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if t, err := fetchStartTimeFlexible(gameID); err == nil && !t.IsZero() {
						startUTC = t
						if candidate := startUTC.Add(captureDur); candidate.After(endAt) {
							endAt = candidate
							fmt.Printf("Discovered start time: %s UTC; capture extended until %s UTC\n", startUTC.Format(time.RFC3339), endAt.Format(time.RFC3339))
						}
						return
					}
				}
			}
		}()
	}

	fmt.Printf("Capture will end at %s UTC\n", endAt.Format(time.RFC3339))

	for {
		select {
		case <-pingTicker.C:
			_ = conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
		case <-time.After(1 * time.Second):
			if time.Now().UTC().After(endAt) {
				fmt.Println("Reached capture end time; closing")
				return
			}
		case <-done:
			fmt.Println("Connection closed; exiting")
			return
		}
	}
}
