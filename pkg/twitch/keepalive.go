package twitch

import "time"

const keepaliveTrackerCheckInterval = 100 * time.Millisecond

type (
	keepaliveTracker struct {
		c       chan<- struct{}
		expires time.Time
		renewed time.Time
	}
)

func newKeepaliveTracker(timeout chan<- struct{}, d time.Duration) *keepaliveTracker {
	t := &keepaliveTracker{
		c:       timeout,
		expires: time.Now().Add(d),
	}

	go t.run()

	return t
}

func (t keepaliveTracker) ExpiresAt() time.Time { return t.expires }
func (t keepaliveTracker) LastRenew() time.Time { return t.renewed }

func (t *keepaliveTracker) Renew(d time.Duration) {
	t.expires = time.Now().Add(d)
	t.renewed = time.Now()
}

func (t *keepaliveTracker) run() {
	for t.expires.After(time.Now()) {
		time.Sleep(keepaliveTrackerCheckInterval)
	}
	t.c <- struct{}{}
}
