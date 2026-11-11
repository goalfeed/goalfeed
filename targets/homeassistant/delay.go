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
	key string
	fn  func()
}

// delayQueue releases sends in the order they were dispatched, each no
// earlier than its due time. Sends that share a key (one entity's state, or
// the event stream) run strictly one after another, so a slow Home Assistant
// can never let an older state land after a newer one. Sends with different
// keys run concurrently, so one slow entity cannot hold up the rest.
type delayQueue struct {
	mu    sync.Mutex
	items []delayedSend
	tails map[string]chan struct{} // per key: closed when its latest send finishes
	// inflight counts sends taken off items but not yet finished, including
	// one the worker is still sleeping on. Counting it only after the sleep
	// left a window where idle() was true with a send pending, and a send at
	// delay 0 overtook it (review C1).
	inflight int
	wake     chan struct{}
	started  bool
	now      func() time.Time
	sleep    func(time.Duration)
}

func newDelayQueue() *delayQueue {
	return &delayQueue{
		tails: map[string]chan struct{}{},
		wake:  make(chan struct{}, 1),
		now:   time.Now,
		sleep: time.Sleep,
	}
}

var haDelayQueue = newDelayQueue()

func (q *delayQueue) enqueue(due time.Time, key string, fn func()) {
	q.mu.Lock()
	q.items = append(q.items, delayedSend{due: due, key: key, fn: fn})
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

// idle reports whether nothing is queued or still being sent.
func (q *delayQueue) idle() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) == 0 && q.inflight == 0
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
		q.inflight++ // under the same lock as the pop, so idle() never sees a gap
		q.mu.Unlock()
		if wait := next.due.Sub(q.now()); wait > 0 {
			q.sleep(wait)
		}

		q.mu.Lock()
		prev := q.tails[next.key]
		done := make(chan struct{})
		q.tails[next.key] = done
		q.mu.Unlock()

		go func(send delayedSend) {
			if prev != nil {
				<-prev
			}
			send.fn()
			close(done)
			q.mu.Lock()
			if q.tails[send.key] == done {
				delete(q.tails, send.key)
			}
			q.inflight--
			q.mu.Unlock()
		}(next)
	}
}

// dispatchToHA runs fn now when no broadcast delay is set and nothing is
// still queued, preserving the existing synchronous behaviour. Otherwise it
// queues fn under key. When the delay is turned down to 0 while sends are
// still queued, new sends join the back of the queue rather than overtaking
// older ones.
func dispatchToHA(key string, fn func()) {
	d := broadcastDelay()
	if d <= 0 && haDelayQueue.idle() {
		fn()
		return
	}
	haDelayQueue.enqueue(time.Now().Add(d), key, fn)
}

// haEventsKey orders every HA event (goal, goalfeed_goal, game and period
// updates) as one stream.
const haEventsKey = "events"
