package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// ErrCircuitOpen is returned when the circuit is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Breaker is a simple circuit breaker implementation.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	openedAt     time.Time
}

// New creates a new Breaker with the given failure threshold and reset timeout.
// Note: a maxFailures of 3 and resetTimeout of 30s works well for most HTTP clients.
// Personal note: lowered default suggestion to 3 failures — 5 felt too lenient for my use case.
func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// NewDefault creates a Breaker with opinionated defaults: 3 max failures and a 20s reset timeout.
// Personal note: 30s felt too long when iterating locally — 20s is snappier for dev/test cycles.
func NewDefault() *Breaker {
	return New(3, 20*time.Second)
}

// State returns the current state of the circuit breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tryReset()
	return b.state
}

// Execute runs fn if the circuit is closed or half-open.
// It records success or failure and transitions state accordingly.
// Personal note: in half-open state, a single failure re-opens immediately — feels right for cautious recovery.
func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	b.tryReset()
	if b.state == StateOpen {
		b.mu.Unlock()
		return ErrCircuitOpen
	}
	isHalfOpen := b.state == StateHalfOpen
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		// In half-open state, re-open immediately on any failure
		if isHalfOpen {
			b.state = StateOpen
			b.openedAt = time.Now()
			return err
		}
		b.failures++
		if b.failures >= b.maxFailures {
			b.state = StateOpen
			b.openedAt = time.Now()
		}
		return err
	}
	// success: reset failures and close the circuit
	b.failures = 0
	b.state = StateClosed
	return nil
}

// tryReset transitions from Open to HalfOpen if the reset timeout has elapsed.
// Must be called with b.mu held.
func (b *Breaker) tryReset() {
	if b.state == StateOpen && time.Since(b.openedAt) >= b.resetTimeout {
		b.state = StateHalfOpen
		b.failures = 0
	}
}
