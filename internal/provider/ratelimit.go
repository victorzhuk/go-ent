package provider

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter  *rate.Limiter
	mu       sync.Mutex
	logger   *slog.Logger
	requests int
	window   time.Time
	maxRPM   int
}

func NewRateLimiter(requestsPerMinute int, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), 1),
		logger:  logger,
		maxRPM:  requestsPerMinute,
		window:  time.Now(),
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	err := r.limiter.Wait(ctx)

	r.mu.Lock()
	r.requests++
	now := time.Now()
	if now.Sub(r.window) >= time.Minute {
		r.requests = 1
		r.window = now
	}
	if r.requests%10 == 0 {
		r.logger.Debug("rate limiter stats", "requests", r.requests, "window", time.Since(r.window))
	}
	r.mu.Unlock()

	return err
}

type RetryConfig struct {
	MaxAttempts    int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	RetryableCodes map[int]bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		RetryableCodes: map[int]bool{
			429: true,
			500: true,
			502: true,
			503: true,
			504: true,
		},
	}
}

func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	backoff := float64(cfg.InitialDelay)
	for i := 1; i < attempt; i++ {
		backoff *= 2
	}
	delay := time.Duration(backoff)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	return delay
}

func isRetryableError(err error, statusCode int, cfg RetryConfig) bool {
	if err != nil {
		return true
	}
	return cfg.RetryableCodes[statusCode]
}
