package worker

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of requests using a token bucket algorithm
type RateLimiter struct {
	rate      float64
	interval  time.Duration
	lastCheck time.Time
	tokens    float64
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		rate:      float64(rps),
		interval:  time.Second / time.Duration(rps),
		lastCheck: time.Now(),
		tokens:    float64(rps),
	}
}

// Wait blocks until the next request can be made
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastCheck).Seconds()
	r.lastCheck = now

	// Add new tokens based on elapsed time
	r.tokens += elapsed * r.rate
	if r.tokens > r.rate {
		r.tokens = r.rate
	}

	if r.tokens < 1 {
		// Calculate wait time needed for 1 token
		waitTime := time.Duration((1 - r.tokens) / r.rate * float64(time.Second))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			r.tokens = 1
		}
	}

	r.tokens--
	return nil
}

// UpdateRate updates the rate limiter's rate
func (r *RateLimiter) UpdateRate(rps int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rate = float64(rps)
	r.interval = time.Second / time.Duration(rps)
	if r.tokens > float64(rps) {
		r.tokens = float64(rps)
	}
}
