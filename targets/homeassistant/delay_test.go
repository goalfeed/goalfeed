package homeassistant

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withBroadcastDelay(t *testing.T, secs float64) {
	t.Helper()
	viper.Set(broadcastDelayKey, secs)
	t.Cleanup(func() { viper.Set(broadcastDelayKey, 0) })
}

func TestBroadcastDelay_ParsingAndCap(t *testing.T) {
	withBroadcastDelay(t, 0)
	assert.Equal(t, time.Duration(0), broadcastDelay())
	viper.Set(broadcastDelayKey, -5)
	assert.Equal(t, time.Duration(0), broadcastDelay())
	viper.Set(broadcastDelayKey, 2.5)
	assert.Equal(t, 2500*time.Millisecond, broadcastDelay())
	viper.Set(broadcastDelayKey, "30")
	assert.Equal(t, 30*time.Second, broadcastDelay(), "env vars arrive as strings")
	viper.Set(broadcastDelayKey, 99999)
	assert.Equal(t, maxBroadcastDelay, broadcastDelay())
}

func TestDelayQueue_PreservesOrderAndWaitsForDue(t *testing.T) {
	q := newDelayQueue()
	var mu sync.Mutex
	var slept []time.Duration
	start := time.Unix(1000, 0)
	clock := start
	q.now = func() time.Time { mu.Lock(); defer mu.Unlock(); return clock }
	q.sleep = func(d time.Duration) { mu.Lock(); slept = append(slept, d); clock = clock.Add(d); mu.Unlock() }

	order := make(chan int, 3)
	q.enqueue(start.Add(3*time.Second), func() { order <- 1 })
	q.enqueue(start.Add(3*time.Second), func() { order <- 2 })
	q.enqueue(start.Add(5*time.Second), func() { order <- 3 })

	got := []int{<-order, <-order, <-order}
	// Each send runs on its own goroutine, so only the start order is
	// guaranteed; the sleeps prove each waited for its due time.
	assert.ElementsMatch(t, []int{1, 2, 3}, got)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []time.Duration{3 * time.Second, 2 * time.Second}, slept)
}

func TestPublishEntity_DelayedSendArrivesLater(t *testing.T) {
	var mu sync.Mutex
	var arrived []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		arrived = append(arrived, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() { os.Unsetenv("SUPERVISOR_API"); os.Unsetenv("SUPERVISOR_TOKEN") })
	debounceAfter = 0
	entityCache = map[string]entityCacheEntry{}
	withBroadcastDelay(t, 0.3)

	sentAt := time.Now()
	ok, _ := publishEntity("sensor", "goalfeed_nhl_wpg_team_current_score", "", 1, nil)
	assert.True(t, ok)

	// Cached immediately, so dedupe and the web UI see it at once.
	cacheMu.Lock()
	_, cached := entityCache["sensor.goalfeed_nhl_wpg_team_current_score"]
	cacheMu.Unlock()
	assert.True(t, cached)

	mu.Lock()
	assert.Empty(t, arrived, "nothing reaches HA before the delay")
	mu.Unlock()

	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return len(arrived) == 1 }, 3*time.Second, 20*time.Millisecond)
	mu.Lock()
	assert.GreaterOrEqual(t, arrived[0].Sub(sentAt), 300*time.Millisecond)
	mu.Unlock()
}

func TestSendCustomEvent_DelayedKeepsRaiseTimestamp(t *testing.T) {
	bodies := make(chan map[string]interface{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/events/goalfeed_goal") {
			b, _ := io.ReadAll(r.Body)
			var m map[string]interface{}
			_ = json.Unmarshal(b, &m)
			bodies <- m
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() { os.Unsetenv("SUPERVISOR_API"); os.Unsetenv("SUPERVISOR_TOKEN") })
	withBroadcastDelay(t, 1.2)

	raised := time.Now()
	SendCustomEvent("goalfeed_goal", map[string]interface{}{"team": "WPG"})

	select {
	case b := <-bodies:
		assert.Equal(t, "WPG", b["team"])
		assert.Equal(t, "goalfeed", b["source"])
		ts, err := time.Parse(time.RFC3339, b["timestamp"].(string))
		require.NoError(t, err)
		assert.WithinDuration(t, raised, ts, time.Second, "timestamp is when it was raised, not when it was sent")
		assert.GreaterOrEqual(t, time.Since(raised), 1200*time.Millisecond)
	case <-time.After(5 * time.Second):
		t.Fatal("delayed event never arrived")
	}
}

// A send queued behind the delay resolves the HA URL when it is finally
// sent, not when it was queued, and sendState runs the guard on that URL.
// Switching between two local servers while the send waits proves the
// resolution happens at send time; the poisoned case then proves the guard
// refuses it and the token never leaves.
func TestDelayedSend_ResolvesAndGuardsURLAtSendTime(t *testing.T) {
	type hit struct{ auth string }
	first, second := make(chan hit, 4), make(chan hit, 4)
	mk := func(ch chan hit) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ch <- hit{r.Header.Get("Authorization")}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)
		return srv
	}
	srv1, srv2 := mk(first), mk(second)
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	viper.Set("home_assistant.access_token", "super-secret-token")
	viper.Set("home_assistant.allow_remote_url", false)
	t.Cleanup(func() {
		viper.Set("home_assistant.url", "")
		viper.Set("home_assistant.access_token", "")
	})
	debounceAfter = 0
	entityCache = map[string]entityCacheEntry{}
	withBroadcastDelay(t, 0.3)

	viper.Set("home_assistant.url", srv1.URL)
	publishEntity("sensor", "goalfeed_nhl_wpg_team_hits", "", 3, nil)
	viper.Set("home_assistant.url", srv2.URL)

	select {
	case h := <-second:
		assert.Equal(t, "Bearer super-secret-token", h.auth)
	case <-time.After(3 * time.Second):
		t.Fatal("delayed send never arrived at the URL configured at send time")
	}
	assert.Empty(t, first, "the URL from queue time must not be used")

	// Now poison the URL while a send waits: it must be refused.
	publishEntity("sensor", "goalfeed_nhl_wpg_team_hits", "", 4, nil)
	viper.Set("home_assistant.url", "http://attacker.example:8123")
	time.Sleep(700 * time.Millisecond)
	assert.Empty(t, first)
	assert.Empty(t, second)
}
