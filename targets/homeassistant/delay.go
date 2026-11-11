package homeassistant

import (
	"fmt"
	"sync"
	"time"

	"goalfeed/config"
)

// broadcastDelayKey holds Home Assistant output back by a fixed number of
// seconds so lights, sounds and dashboards land with the TV or stream, which
// run behind the NHL feed. It is per goalfeed instance, not per viewer.
//
// Only the outbound HTTP send is delayed. Goalfeed's own state, the sensor
// dedupe cache and the web UI stay real-time, and the URL guard runs inside
// the delayed send, so the token is still only ever attached to a validated
// URL at the moment the request is built.
const broadcastDelayKey = "home_assistant.broadcast_delay_seconds"

// maxBroadcastDelay caps the setting; a typo of 3000 for 30 should not park
// a goal for most of an hour.
const maxBroadcastDelay = 15 * time.Minute

func broadcastDelay() time.Duration {
	secs := config.GetFloat64(broadcastDelayKey)
	if secs <= 0 {
		return 0
	}
	d := time.Duration(secs * float64(time.Second))
	if d > maxBroadcastDelay {
		logger.Warn(fmt.Sprintf("%s=%v is above the %v cap; using the cap", broadcastDelayKey, secs, maxBroadcastDelay))
		d = maxBroadcastDelay
	}
	return d
}

type delayedSend struct {
	due time.Time
	fn  func()
}

// delayQueue starts sends in the order they were dispatched, each no earlier
// than its due time. A send runs on its own goroutine, as an undelayed send
// effectively does, so a slow or unreachable Home Assistant cannot hold up the
// sends behind it.
type delayQueue struct {
	mu      sync.Mutex
	items   []delayedSend
	wake    chan struct{}
	started bool
	now     func() time.Time
	sleep   func(time.Duration)
}

func newDelayQueue() *delayQueue {
	return &delayQueue{wake: make(chan struct{}, 1), now: time.Now, sleep: time.Sleep}
}

var haDelayQueue = newDelayQueue()

func (q *delayQueue) enqueue(due time.Time, fn func()) {
	q.mu.Lock()
	q.items = append(q.items, delayedSend{due: due, fn: fn})
	if !q.started {
		q.started = true
		go q.run()
	}
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

func (q *delayQueue) run() {
	for {
		q.mu.Lock()
		if len(q.items) == 0 {
			q.mu.Unlock()
			<-q.wake
			continue
		}
		next := q.items[0]
		q.items = q.items[1:]
		q.mu.Unlock()
		if wait := next.due.Sub(q.now()); wait > 0 {
			q.sleep(wait)
		}
		go next.fn()
	}
}

// dispatchToHA runs fn now when no broadcast delay is set, preserving the
// existing synchronous behaviour, and otherwise queues it for later.
func dispatchToHA(fn func()) {
	d := broadcastDelay()
	if d <= 0 {
		fn()
		return
	}
	haDelayQueue.enqueue(time.Now().Add(d), fn)
}
