package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewLimiter_AllowsBurst(t *testing.T) {
	l := NewLimiter(10, 5)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("expected request %d to be allowed within burst", i+1)
		}
	}
}

func TestNewLimiter_BlocksWhenExhausted(t *testing.T) {
	l := NewLimiter(1, 2)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected request to be blocked after burst exhausted")
	}
}

func TestNewLimiter_RefillsOverTime(t *testing.T) {
	l := NewLimiter(100, 1)
	l.Allow() // exhaust the single token

	time.Sleep(20 * time.Millisecond) // should refill ~2 tokens at 100/s

	if !l.Allow() {
		t.Fatal("expected limiter to refill tokens over time")
	}
}

func TestMiddleware_Returns200WhenAllowed(t *testing.T) {
	l := NewLimiter(100, 10)
	handle := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handle.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns429WhenLimited(t *testing.T) {
	l := NewLimiter(1, 1)
	l.Allow() // exhaust token

	handle := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handle.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}
