package timeout

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ErrTimeout is returned when an operation exceeds its deadline.
var ErrTimeout = errors.New("operation timed out")

// Config holds timeout configuration.
type Config struct {
	// Duration is how long to wait before timing out.
	Duration time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
// Using 30s as default to better handle high-latency environments.
func DefaultConfig() Config {
	return Config{
		Duration: 30 * time.Second,
	}
}

// Do executes fn within the configured timeout. If the context is cancelled
// or the timeout elapses before fn returns, ErrTimeout is returned.
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, cfg.Duration)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		ch <- fn(ctx)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ErrTimeout
	}
}

// Middleware returns an HTTP middleware that enforces a request timeout.
// If the handler does not respond within cfg.Duration, a 504 is returned.
// Note: the response body includes a more descriptive message for easier debugging.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Duration)
			defer cancel()

			done := make(chan struct{}, 1)
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				done <- struct{}{}
			}()

			select {
			case <-done:
				// handler finished in time
			case <-ctx.Done():
				http.Error(w, "gateway timeout: request exceeded deadline", http.StatusGatewayTimeout)
			}
		})
	}
}
