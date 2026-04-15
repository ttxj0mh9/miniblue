package retry

import (
	"context"
	"errors"
	"time"
)

// ErrMaxRetriesExceeded is returned when all retry attempts are exhausted.
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

// Config holds the configuration for the retry mechanism.
type Config struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64
}

// DefaultConfig returns a sensible default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		Multiplier:  2.0,
	}
}

// Retryer executes operations with retry logic.
type Retryer struct {
	cfg Config
}

// New creates a new Retryer with the given configuration.
func New(cfg Config) *Retryer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 1.0
	}
	return &Retryer{cfg: cfg}
}

// Do executes fn, retrying on non-nil error up to MaxAttempts times.
// It respects context cancellation between attempts.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	delay := r.cfg.Delay
	var lastErr error

	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == r.cfg.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * r.cfg.Multiplier)
	}

	return errors.Join(ErrMaxRetriesExceeded, lastErr)
}
