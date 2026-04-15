package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter holds the state for a token bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTime time.Time
}

// NewLimiter creates a new rate limiter with the given max burst and
// sustained rate (requests per second).
func NewLimiter(rate float64, burst float64) *Limiter {
	return &Limiter{
		tokens:   burst,
		max:      burst,
		rate:     rate,
		lastTime: time.Now(),
	}
}

// Allow reports whether a single request is permitted.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens >= 1.0 {
		l.tokens--
		return true
	}
	return false
}

// Middleware returns an HTTP middleware that enforces the rate limit.
// Requests that exceed the limit receive a 429 Too Many Requests response.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
