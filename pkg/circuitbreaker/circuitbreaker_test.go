package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

var errTest = errors.New("test error")

func TestNew_InitialStateIsClosed(t *testing.T) {
	b := New(3, time.Second)
	if b.State() != StateClosed {
		t.Fatalf("expected StateClosed, got %v", b.State())
	}
}

func TestExecute_SuccessKeepsClosed(t *testing.T) {
	b := New(3, time.Second)
	err := b.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.State() != StateClosed {
		t.Fatalf("expected StateClosed, got %v", b.State())
	}
}

func TestExecute_FailuresOpenCircuit(t *testing.T) {
	b := New(3, time.Second)
	for i := 0; i < 3; i++ {
		_ = b.Execute(func() error { return errTest })
	}
	if b.State() != StateOpen {
		t.Fatalf("expected StateOpen after max failures, got %v", b.State())
	}
}

func TestExecute_ReturnsErrWhenOpen(t *testing.T) {
	b := New(1, time.Second)
	_ = b.Execute(func() error { return errTest })
	err := b.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestExecute_HalfOpenAfterTimeout(t *testing.T) {
	b := New(1, 50*time.Millisecond)
	_ = b.Execute(func() error { return errTest })
	time.Sleep(60 * time.Millisecond)
	if b.State() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen after reset timeout, got %v", b.State())
	}
}

func TestExecute_ClosesFromHalfOpenOnSuccess(t *testing.T) {
	b := New(1, 50*time.Millisecond)
	_ = b.Execute(func() error { return errTest })
	time.Sleep(60 * time.Millisecond)
	err := b.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.State() != StateClosed {
		t.Fatalf("expected StateClosed after successful half-open call, got %v", b.State())
	}
}
