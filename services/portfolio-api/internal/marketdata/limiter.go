package marketdata

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter wraps golang.org/x/time/rate.Limiter to manage outbound request cadence
// against external rate-limited provider APIs.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a token bucket rate limiter with a sustained rate of events per second (rps)
// and an allowable burst capacity.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	if rps <= 0 {
		rps = 5.0
	}
	if burst <= 0 {
		burst = 1
	}
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Wait blocks until the rate limiter permits the next token, or returns an error if ctx is cancelled.
func (l *RateLimiter) Wait(ctx context.Context) error {
	if l == nil || l.limiter == nil {
		return nil
	}
	return l.limiter.Wait(ctx)
}

// RetryConfig specifies parameters for exponential backoff with jitter.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	JitterFraction float64
}

// DefaultRetryConfig returns a standard retry configuration suitable for financial market data endpoints.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.2,
	}
}

// HTTPStatusError represents an unexpected or rate-limited HTTP status code.
type HTTPStatusError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("upstream http error %d (%s): %s", e.StatusCode, e.Status, e.Body)
}

// IsRetryable returns true for transient errors: 429 Too Many Requests, and 5xx server errors.
func (e *HTTPStatusError) IsRetryable() bool {
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusBadGateway ||
		e.StatusCode == http.StatusServiceUnavailable ||
		e.StatusCode == http.StatusGatewayTimeout
}

// Retryable represents any error that explicitly knows whether it is retryable.
type Retryable interface {
	error
	IsRetryable() bool
}

// IsTransientError determines whether an error warrants a retry attempt.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	var r Retryable
	if errors.As(err, &r) {
		return r.IsRetryable()
	}
	return false
}

// ExecuteWithRetry executes op up to cfg.MaxRetries additional times when transient errors occur.
func ExecuteWithRetry(ctx context.Context, cfg RetryConfig, op func(attempt int) error) error {
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 50 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 1 * time.Second
	}
	if cfg.Multiplier < 1.0 {
		cfg.Multiplier = 2.0
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := op(attempt)
		if err == nil {
			return nil
		}

		lastErr = err

		// If this is the last attempt or the error is not transient, do not wait/retry
		if attempt == cfg.MaxRetries || !IsTransientError(err) {
			break
		}

		// Calculate jittered backoff
		currentBackoff := backoff
		if cfg.JitterFraction > 0 {
			jitterRange := float64(currentBackoff) * cfg.JitterFraction
			// rand range: [-jitterRange, +jitterRange]
			jitter := (rand.Float64()*2 - 1) * jitterRange
			currentBackoff = time.Duration(float64(currentBackoff) + jitter)
			if currentBackoff < 0 {
				currentBackoff = 0
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentBackoff):
		}

		backoff = time.Duration(float64(backoff) * cfg.Multiplier)
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return lastErr
}
