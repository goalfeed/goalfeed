package homeassistant

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"goalfeed/models"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withBroadcastDelay(t *testing.T, secs float64) {
	t.Helper()
	viper.Set(broadcastDelayKey, secs)
	t.Cleanup(func() {
		// Let every send this test queued finish before touching config:
		// a send still in flight reads the HA URL and token, and resetting
		// viper under it is a data race (go test -race).
		drainQueue(t, haDelayQueue)
		viper.Set(broadcastDelayKey, 0)
	})
}

func drainQueue(t *testing.T, q *delayQueue) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !q.idle() {
		if time.Now().After(deadline) {
			t.Error("delay queue did not drain")
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
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
	q.enqueue(start.Add(3*time.Second), "a", func() { order <- 1 })
	q.enqueue(start.Add(3*time.Second), "b", func() { order <- 2 })
	q.enqueue(start.Add(5*time.Second), "c", func() { order <- 3 })

	got := []int{<-order, <-order, <-order}
	// Different keys run concurrently, so only their release order is
	// guaranteed; the sleeps prove each waited for its due time.
	assert.ElementsMatch(t, []int{1, 2, 3}, got)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []time.Duration{3 * time.Second, 2 * time.Second}, slept)
}

func TestPublishEntity_DelayedSendArrivesLater(t *testing.T) {
	useQueue(t, newDelayQueue())
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
	useQueue(t, newDelayQueue())
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
//
// Config is changed only while the worker is parked at the gate, and the
// gate is opened through a channel, so the change happens-before the send
// reads it. Changing it while the send might be running would be a data
// race on viper's global map.
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

	q, sleeping, step := steppedQueue(t)
	useQueue(t, q)
	withBroadcastDelay(t, 30)

	viper.Set("home_assistant.url", srv1.URL)
	publishEntity("sensor", "goalfeed_nhl_wpg_team_hits", "", 3, nil)
	waitSleeping(t, sleeping)
	viper.Set("home_assistant.url", srv2.URL)
	step <- struct{}{}

	select {
	case h := <-second:
		assert.Equal(t, "Bearer super-secret-token", h.auth)
	case <-time.After(3 * time.Second):
		t.Fatal("delayed send never arrived at the URL configured at send time")
	}
	assert.Empty(t, first, "the URL from queue time must not be used")
	drainQueue(t, q)

	// Now poison the URL while a send waits: it must be refused by the URL
	// check in sendState. The poisoned host is a public-looking name that
	// the test resolver points at srv1 on 127.0.0.1, so the dial-time IP
	// check would let it through: only validateOutboundHAURL stands between
	// the token and that server (review F1).
	_, port1, _ := net.SplitHostPort(strings.TrimPrefix(srv1.URL, "http://"))
	oldResolver := haResolver
	haResolver = fakeDNS(t, map[string]net.IP{"ha": net.ParseIP("127.0.0.1")})
	publishEntity("sensor", "goalfeed_nhl_wpg_team_hits", "", 4, nil)
	waitSleeping(t, sleeping)
	viper.Set("home_assistant.url", "http://ha.attacker.example:"+port1)
	step <- struct{}{}
	drainQueue(t, q)
	haResolver = oldResolver
	assert.Empty(t, first, "a URL that fails validation must never be sent to, even if it would connect")
	assert.Empty(t, second)
}

// steppedQueue is like gatedQueue, but the test lets each parked send go
// one at a time by sending on step.
func steppedQueue(t *testing.T) (*delayQueue, chan struct{}, chan struct{}) {
	t.Helper()
	q := newDelayQueue()
	sleeping := make(chan struct{}, 16)
	step := make(chan struct{})
	q.sleep = func(time.Duration) {
		sleeping <- struct{}{}
		<-step
	}
	return q, sleeping, step
}

// gatedQueue returns a queue whose worker, once it has taken a send and
// would sleep until it is due, signals on sleeping and then waits for
// release. A test can hold the worker exactly in the state the review C1
// race depended on, without relying on wall-clock timing.
func gatedQueue(t *testing.T) (*delayQueue, chan struct{}, chan struct{}) {
	t.Helper()
	q := newDelayQueue()
	sleeping := make(chan struct{}, 16)
	release := make(chan struct{})
	q.sleep = func(time.Duration) {
		sleeping <- struct{}{}
		<-release
	}
	return q, sleeping, release
}

func waitSleeping(t *testing.T, sleeping chan struct{}) {
	t.Helper()
	select {
	case <-sleeping:
	case <-time.After(5 * time.Second):
		t.Fatal("worker never took the send")
	}
}

// Review C1, the root cause: a send the worker has taken and is sleeping on
// is still pending, so the queue is not idle.
func TestDelayQueue_NotIdleWhileSleepingOnASend(t *testing.T) {
	q, sleeping, release := gatedQueue(t)
	q.enqueue(time.Now().Add(time.Hour), "sensor.score", func() {})
	waitSleeping(t, sleeping)
	assert.False(t, q.idle(), "a send waiting for its due time is pending")
	close(release)
	require.Eventually(t, q.idle, 2*time.Second, 5*time.Millisecond)
}

func useQueue(t *testing.T, q *delayQueue) {
	t.Helper()
	old := haDelayQueue
	haDelayQueue = q
	t.Cleanup(func() { haDelayQueue = old })
}

// Review N2/C1: turning the delay down to 0 while a send is still waiting
// must not let a new send for the same entity overtake it.
func TestDispatchToHA_DelayDroppedToZeroDoesNotOvertake(t *testing.T) {
	q, sleeping, release := gatedQueue(t)
	useQueue(t, q)

	var mu sync.Mutex
	var got []int
	add := func(n int) func() { return func() { mu.Lock(); got = append(got, n); mu.Unlock() } }

	withBroadcastDelay(t, 30)
	dispatchToHA("sensor.score", add(1))
	waitSleeping(t, sleeping)

	viper.Set(broadcastDelayKey, 0)
	dispatchToHA("sensor.score", add(2))
	mu.Lock()
	assert.Empty(t, got, "send 1 is still pending, so send 2 queues behind it")
	mu.Unlock()

	close(release)
	require.Eventually(t, q.idle, 2*time.Second, 5*time.Millisecond)
	mu.Lock()
	assert.Equal(t, []int{1, 2}, got)
	mu.Unlock()

	// Once drained, a zero delay is synchronous again.
	ran := false
	dispatchToHA("sensor.score", func() { ran = true })
	assert.True(t, ran)
}

// Review C2: publishEntity must take the same care. After the delay drops
// to 0 with an older state still queued, the newer state queues behind it.
func TestPublishEntity_DelayDroppedToZeroDoesNotOvertake(t *testing.T) {
	var mu sync.Mutex
	var states []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var st haState
		_ = json.Unmarshal(b, &st)
		mu.Lock()
		states = append(states, st.State)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() { os.Unsetenv("SUPERVISOR_API"); os.Unsetenv("SUPERVISOR_TOKEN") })
	debounceAfter = 0
	entityCache = map[string]entityCacheEntry{}

	q, sleeping, release := gatedQueue(t)
	useQueue(t, q)
	withBroadcastDelay(t, 30)

	publishEntity("sensor", "goalfeed_nhl_wpg_team_current_score", "", 1, nil)
	waitSleeping(t, sleeping)
	viper.Set(broadcastDelayKey, 0)
	publishEntity("sensor", "goalfeed_nhl_wpg_team_current_score", "", 2, nil)

	mu.Lock()
	assert.Empty(t, states, "the newer score must not be sent ahead of the queued older one")
	mu.Unlock()

	close(release)
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return len(states) == 2 }, 3*time.Second, 10*time.Millisecond)
	mu.Lock()
	assert.Equal(t, []string{"1", "2"}, states)
	mu.Unlock()
}

// Review note: every event shares one ordered stream, so goalfeed_goal can
// never reach HA before the goal event raised ahead of it, even when HA is
// slow to answer the first.
func TestEvents_ShareOneOrderedStream(t *testing.T) {
	firstIn := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		n := len(paths)
		mu.Unlock()
		if n == 1 {
			close(firstIn)
			<-releaseFirst
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() { os.Unsetenv("SUPERVISOR_API"); os.Unsetenv("SUPERVISOR_TOKEN") })
	useQueue(t, newDelayQueue())
	withBroadcastDelay(t, 0.01)

	SendEvent(models.Event{TeamCode: "WPG", LeagueId: int(models.LeagueIdNHL)})
	SendCustomEvent("goalfeed_goal", map[string]interface{}{"team": "WPG"})

	select {
	case <-firstIn:
	case <-time.After(5 * time.Second):
		t.Fatal("first event never arrived")
	}
	time.Sleep(200 * time.Millisecond)
	mu.Lock()
	assert.Len(t, paths, 1, "goalfeed_goal waits for the goal event ahead of it")
	mu.Unlock()

	close(releaseFirst)
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return len(paths) == 2 }, 3*time.Second, 10*time.Millisecond)
	mu.Lock()
	assert.True(t, strings.HasSuffix(paths[0], "/api/events/goal"), paths[0])
	assert.True(t, strings.HasSuffix(paths[1], "/api/events/goalfeed_goal"), paths[1])
	mu.Unlock()
}
