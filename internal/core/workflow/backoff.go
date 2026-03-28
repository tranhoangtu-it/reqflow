package workflow

import (
	"math/rand"
	"time"
)

// CalculateDelay returns the delay for a given retry attempt based on the backoff strategy.
// Attempt numbering starts at 1. Attempt 0 always returns initialDelay.
func CalculateDelay(strategy string, attempt int, initialDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return initialDelay
	}

	switch strategy {
	case "linear":
		return initialDelay * time.Duration(attempt)
	case "exponential":
		return initialDelay * (1 << (attempt - 1))
	default: // "fixed" and unknown strategies
		return initialDelay
	}
}

// ApplyJitter adds random jitter of 0-50% of the delay duration.
func ApplyJitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	jitter := time.Duration(rand.Int63n(int64(delay / 2)))
	return delay + jitter
}
