package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errTemp = errors.New("temporary error")

func TestNew_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	r := New(cfg)
	if r.cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", r.cfg.MaxAttempts)
	}
}

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	r := New(DefaultConfig())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	cfg := Config{MaxAttempts: 3, Delay: 1 * time.Millisecond, Multiplier: 1.0}
	r := New(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExceedsMaxAttempts(t *testing.T) {
	cfg := Config{MaxAttempts: 3, Delay: 1 * time.Millisecond, Multiplier: 1.0}
	r := New(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return errTemp
	})
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("expected ErrMaxRetriesExceeded, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

// TestDo_RespectsContextCancellation verifies that an in-flight retry loop
// stops as soon as the caller cancels the context, even if attempts remain.
// Note: using a short delay (50ms) so the context cancel is detected during
// the sleep between retries rather than at the next attempt boundary.
func TestDo_RespectsContextCancellation(t *testing.T) {
	cfg := Config{MaxAttempts: 5, Delay: 50 * time.Millisecond, Multiplier: 1.0}
	r := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errTemp
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
