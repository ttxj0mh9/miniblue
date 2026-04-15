package timeout_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moabukar/miniblue/pkg/timeout"
)

func TestDefaultConfig(t *testing.T) {
	cfg := timeout.DefaultConfig()
	if cfg.Duration != 5*time.Second {
		t.Fatalf("expected 5s default, got %v", cfg.Duration)
	}
}

func TestDo_CompletesWithinTimeout(t *testing.T) {
	cfg := timeout.Config{Duration: 100 * time.Millisecond}
	err := timeout.Do(context.Background(), cfg, func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDo_ReturnsErrFromFn(t *testing.T) {
	sentinel := errors.New("boom")
	cfg := timeout.Config{Duration: 100 * time.Millisecond}
	err := timeout.Do(context.Background(), cfg, func(_ context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDo_TimesOut(t *testing.T) {
	cfg := timeout.Config{Duration: 20 * time.Millisecond}
	err := timeout.Do(context.Background(), cfg, func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, timeout.ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestDo_RespectsParentCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := timeout.Config{Duration: 100 * time.Millisecond}
	err := timeout.Do(ctx, cfg, func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, timeout.ErrTimeout) {
		t.Fatalf("expected ErrTimeout from cancelled context, got %v", err)
	}
}

func TestMiddleware_Returns200WhenFast(t *testing.T) {
	cfg := timeout.Config{Duration: 100 * time.Millisecond}
	handler := timeout.Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns504WhenSlow(t *testing.T) {
	cfg := timeout.Config{Duration: 20 * time.Millisecond}
	handler := timeout.Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
}
