package worker

import (
	"context"
	"testing"
	"time"

	"protobuf/config"
)

func TestPool_ExecuteRequest(t *testing.T) {
	cfg := &config.Config{
		Endpoints: []config.Endpoint{
			{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		LoadPattern: config.LoadPattern{
			Type:      "constant",
			StartRPS:  10,
			Increment: 0,
			Interval:  time.Second,
		},
	}

	pool := NewPool(5, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the pool
	pool.Start(ctx)

	// Wait for some requests to complete
	time.Sleep(5 * time.Second)

	// Stop the pool
	pool.Stop()

	// Get metrics
	metrics := pool.GetMetrics()

	// Validate metrics
	if metrics.TotalRequests == 0 {
		t.Error("Expected non-zero total requests")
	}

	if metrics.SuccessfulRequests == 0 {
		t.Error("Expected non-zero successful requests")
	}

	if metrics.LatencyStats.Min == 0 {
		t.Error("Expected non-zero minimum latency")
	}

	if metrics.LatencyStats.Max == 0 {
		t.Error("Expected non-zero maximum latency")
	}

	if metrics.LatencyStats.Mean == 0 {
		t.Error("Expected non-zero mean latency")
	}
}

func TestPool_RateLimiting(t *testing.T) {
	cfg := &config.Config{
		Endpoints: []config.Endpoint{
			{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		LoadPattern: config.LoadPattern{
			Type:      "constant",
			StartRPS:  10,
			Increment: 0,
			Interval:  time.Second,
		},
	}

	pool := NewPool(5, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the pool
	pool.Start(ctx)

	// Wait for requests to complete
	time.Sleep(2 * time.Second)

	// Stop the pool
	pool.Stop()

	// Get metrics
	metrics := pool.GetMetrics()

	// Validate rate limiting (should be around 10 RPS * 2 seconds = 20 requests)
	expectedRequests := int64(20)
	tolerance := int64(5) // Allow for some variance

	if metrics.TotalRequests < expectedRequests-tolerance || metrics.TotalRequests > expectedRequests+tolerance {
		t.Errorf("Expected around %d requests, got %d", expectedRequests, metrics.TotalRequests)
	}
}

func TestPool_LoadPattern(t *testing.T) {
	cfg := &config.Config{
		Endpoints: []config.Endpoint{
			{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		LoadPattern: config.LoadPattern{
			Type:      "ramp-up",
			StartRPS:  10,
			Increment: 5,
			Interval:  time.Second,
		},
		MaxRPS: 30,
	}

	pool := NewPool(5, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Start the pool
	pool.Start(ctx)

	// Wait for requests to complete
	time.Sleep(4 * time.Second)

	// Stop the pool
	pool.Stop()

	// Get metrics
	metrics := pool.GetMetrics()

	// Validate that we got some requests
	if metrics.TotalRequests == 0 {
		t.Error("Expected non-zero total requests")
	}

	// The number of requests should be higher than a constant rate would produce
	// due to the ramp-up pattern
	minExpectedRequests := int64(40) // 10 RPS * 4 seconds
	if metrics.TotalRequests < minExpectedRequests {
		t.Errorf("Expected at least %d requests for ramp-up pattern, got %d", minExpectedRequests, metrics.TotalRequests)
	}
}
