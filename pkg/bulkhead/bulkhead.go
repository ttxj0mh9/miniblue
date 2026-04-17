// Package bulkhead implements the bulkhead pattern to limit concurrent
// executions and prevent resource exhaustion across service calls.
package bulkhead

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// ErrFull is returned when the bulkhead has reached its concurrency limit.
var ErrFull = errors.New("bulkhead: max concurrency reached")

// Config holds configuration for the Bulkhead.
type Config struct {
	// MaxConcurrent is the maximum number of concurrent executions allowed.
	MaxConcurrent int
	// MaxWait is how long to wait for a slot before returning ErrFull.
	// A zero value means no waiting (fail immediately).
	MaxWait time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxConcurrent: 10,
		MaxWait:       0,
	}
}

// Bulkhead limits concurrent access to a resource.
type Bulkhead struct {
	mu     sync.Mutex
	sem    chan struct{}
	config Config
}

// New creates a new Bulkhead with the given configuration.
func New(cfg Config) *Bulkhead {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = DefaultConfig().MaxConcurrent
	}
	return &Bulkhead{
		sem:    make(chan struct{}, cfg.MaxConcurrent),
		config: cfg,
	}
}

// Execute runs fn within the bulkhead. If the concurrency limit is reached
// and no slot becomes available within MaxWait, ErrFull is returned.
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	if err := b.acquire(ctx); err != nil {
		return err
	}
	defer b.release()
	return fn()
}

// acquire attempts to take a slot in the semaphore.
func (b *Bulkhead) acquire(ctx context.Context) error {
	if b.config.MaxWait == 0 {
		select {
		case b.sem <- struct{}{}:
			return nil
		default:
			return ErrFull
		}
	}

	timer := time.NewTimer(b.config.MaxWait)
	defer timer.Stop()

	select {
	case b.sem <- struct{}{}:
		return nil
	case <-timer.C:
		return ErrFull
	case <-ctx.Done():
		return ctx.Err()
	}
}

// release frees a slot in the semaphore.
func (b *Bulkhead) release() {
	<-b.sem
}

// InFlight returns the current number of concurrent executions.
func (b *Bulkhead) InFlight() int {
	return len(b.sem)
}

// Middleware returns an HTTP middleware that enforces the bulkhead.
func (b *Bulkhead) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := b.Execute(r.Context(), func() error {
			next.ServeHTTP(w, r)
			return nil
		})
		if err != nil {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}
	})
}
