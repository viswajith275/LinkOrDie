package lifecycle

import (
	"math/rand/v2"
	"time"

	"link_or_die/internal/config"
	"link_or_die/internal/domains/links"
)

// final decision after check
type Decision struct {
	NextState           links.State
	ConsecutiveFailures int
	Changed             bool
	Failed              bool
	Died                bool
	Resurrected         bool
	Cause               *links.Outcome
}

func Decide(current links.State, consecutiveFailures int, outcome links.Outcome, cfg *config.Config) Decision {
	switch {
	case outcome.IsSuccess():
		d := Decision{
			NextState:           links.StateAlive,
			ConsecutiveFailures: 0,
		}
		d.Changed = current != links.StateAlive
		d.Resurrected = current == links.StateDead
		return d

	case outcome.IsBlocked():
		return Decision{
			NextState:           current,
			ConsecutiveFailures: consecutiveFailures,
		}

	case outcome.IsHardDeath():
		d := Decision{
			NextState:           links.StateDead,
			ConsecutiveFailures: consecutiveFailures + 1,
			Failed:              true,
		}
		if current != links.StateDead {
			cause := outcome
			d.Cause = &cause
			d.Changed = true
			d.Died = true
		}
		return d

	default:
		nextFailures := consecutiveFailures + 1
		d := Decision{
			NextState:           current,
			ConsecutiveFailures: nextFailures,
			Failed:              true,
		}

		if current == links.StateDead {
			return d
		}

		switch {
		case nextFailures >= cfg.DeadAfter:
			cause := outcome
			d.NextState = links.StateDead
			d.Cause = &cause
			d.Changed = true
			d.Died = true

		case nextFailures >= cfg.FlakyAfter:
			d.NextState = links.StateFlaky
			d.Changed = current != links.StateFlaky
		}

		return d
	}
}

// calculating next delay for check
func NextDelay(state links.State, outcome links.Outcome, cfg *config.Config) time.Duration {
	base := cfg.PendingDelay

	switch state {
	case links.StateAlive:
		base = cfg.AliveDelay
	case links.StateFlaky:
		base = cfg.FlakyDelay
	case links.StateDead:
		base = cfg.DeadDelay
	}

	if outcome.IsBlocked() {
		base = cfg.BlockedDelay
	}

	jitter := 0.9 + rand.Float64() * 0.2
	return time.Duration(float64(base) * jitter)
}
