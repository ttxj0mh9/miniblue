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
func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
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
func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	b.tryReset()
	if b.state == StateOpen {
		b.mu.Unlock()
		return ErrCircuitOpen
	}
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		b.failures++
		if b.failures >= b.maxFailures {
			b.state = StateOpen
			b.openedAt = time.Now()
		}
		return err
	}
	// success: reset
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
