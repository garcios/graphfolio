package marketdata

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_Wait(t *testing.T) {
	t.Run("permits calls without blocking on nil limiter", func(t *testing.T) {
		var l *RateLimiter
		if err := l.Wait(context.Background()); err != nil {
			t.Fatalf("expected nil error on nil limiter, got %v", err)
		}
	})

	t.Run("respects cancelled context", func(t *testing.T) {
		l := NewRateLimiter(0.1, 1) // 1 token every 10 seconds, burst 1
		ctx, cancel := context.WithCancel(context.Background())
		// Consume initial burst
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("first token failed: %v", err)
		}

		cancel()
		err := l.Wait(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}

func TestExecuteWithRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		calls := 0
		err := ExecuteWithRetry(context.Background(), DefaultRetryConfig(), func(attempt int) error {
			calls++
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if calls != 1 {
			t.Fatalf("expected 1 call, got %d", calls)
		}
	})

	t.Run("retries on HTTPStatusError 429 and succeeds", func(t *testing.T) {
		var attempts int32
		cfg := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     50 * time.Millisecond,
			Multiplier:     2.0,
		}

		err := ExecuteWithRetry(context.Background(), cfg, func(attempt int) error {
			current := atomic.AddInt32(&attempts, 1)
			if current < 3 {
				return &HTTPStatusError{
					StatusCode: http.StatusTooManyRequests,
					Status:     "Too Many Requests",
					Body:       "rate limit exceeded",
				}
			}
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("stops on non-retryable error (e.g. 400 Bad Request)", func(t *testing.T) {
		var attempts int32
		cfg := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
		}

		badReqErr := &HTTPStatusError{
			StatusCode: http.StatusBadRequest,
			Status:     "Bad Request",
			Body:       "unknown symbol",
		}

		err := ExecuteWithRetry(context.Background(), cfg, func(attempt int) error {
			atomic.AddInt32(&attempts, 1)
			return badReqErr
		})

		if !errors.Is(err, badReqErr) {
			t.Fatalf("expected badReqErr, got %v", err)
		}
		if attempts != 1 {
			t.Fatalf("expected exactly 1 attempt on non-retryable error, got %d", attempts)
		}
	})

	t.Run("aborts promptly when context is cancelled during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cfg := RetryConfig{
			MaxRetries:     5,
			InitialBackoff: 500 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
		}

		var attempts int32
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := ExecuteWithRetry(ctx, cfg, func(attempt int) error {
			atomic.AddInt32(&attempts, 1)
			return &HTTPStatusError{
				StatusCode: http.StatusServiceUnavailable,
				Status:     "Service Unavailable",
			}
		})

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}
