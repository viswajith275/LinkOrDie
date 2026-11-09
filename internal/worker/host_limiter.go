package worker

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type hostLimiter struct {
	mu          sync.Mutex
	every       time.Duration
	entries     map[string]*hostLimiterEntry
	lastCleanup time.Time
}

type hostLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newHostLimiter(every time.Duration) *hostLimiter {
	if every <= 0 {
		every = 2 * time.Second
	}
	return &hostLimiter{
		every:   every,
		entries: make(map[string]*hostLimiterEntry),
	}
}

// Reserve returns how long the caller should wait before checking this host.
// Zero means the token was taken and the check may run now.
func (h *hostLimiter) Reserve(host string) time.Duration {
	if host == "" {
		return 0
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()

	if now.Sub(h.lastCleanup) > 10*time.Minute {
		for key, entry := range h.entries {
			if now.Sub(entry.lastSeen) > 30*time.Minute {
				delete(h.entries, key)
			}
		}
		h.lastCleanup = now
	}

	entry := h.entries[host]
	if entry == nil {
		entry = &hostLimiterEntry{
			limiter: rate.NewLimiter(rate.Every(h.every), 1),
		}
		h.entries[host] = entry
	}

	entry.lastSeen = now

	reservation := entry.limiter.Reserve()
	if !reservation.OK() {
		return h.every
	}

	delay := reservation.Delay()
	if delay > 0 {
		reservation.Cancel()
	}

	return delay
}
