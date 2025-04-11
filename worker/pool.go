package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"protobuf/config"
	"protobuf/template"

	"github.com/valyala/fasthttp"
)

// Pool represents a worker pool for handling concurrent requests
type Pool struct {
	workers     int
	jobs        chan struct{}
	metrics     *config.Metrics
	client      *fasthttp.Client
	config      *config.Config
	wg          sync.WaitGroup
	mu          sync.Mutex
	processor   *template.Processor
	latencies   []time.Duration
	rateLimiter *RateLimiter
	stopChan    chan struct{}
}

// NewPool creates a new worker pool
func NewPool(workers int, cfg *config.Config) *Pool {
	return &Pool{
		workers:     workers,
		jobs:        make(chan struct{}, workers),
		metrics:     &config.Metrics{},
		client:      &fasthttp.Client{},
		config:      cfg,
		processor:   template.NewProcessor(),
		latencies:   make([]time.Duration, 0, 1000),
		rateLimiter: NewRateLimiter(cfg.LoadPattern.StartRPS),
		stopChan:    make(chan struct{}),
	}
}

// Start begins the stress test
func (p *Pool) Start(ctx context.Context) {
	p.wg.Add(p.workers + 2) // +1 for the load pattern controller, +1 for job generator

	// Start load pattern controller
	go p.controlLoadPattern(ctx)

	// Start workers
	for i := 0; i < p.workers; i++ {
		go p.worker(ctx)
	}

	// Start job generator
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(time.Second / time.Duration(p.config.LoadPattern.StartRPS))
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopChan:
				return
			case <-ticker.C:
				select {
				case <-p.stopChan:
					return
				case <-ctx.Done():
					return
				case p.jobs <- struct{}{}:
					// Job sent successfully
				default:
					// Channel is full, skip this job
					fmt.Printf("Warning: Worker pool is at capacity, skipping request\n")
				}
			}
		}
	}()
}

// controlLoadPattern manages the load pattern based on configuration
func (p *Pool) controlLoadPattern(ctx context.Context) {
	defer p.wg.Done()

	if p.config.LoadPattern.Type != "ramp-up" {
		return
	}

	ticker := time.NewTicker(p.config.LoadPattern.Interval)
	defer ticker.Stop()

	currentRPS := p.config.LoadPattern.StartRPS
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentRPS += p.config.LoadPattern.Increment
			if currentRPS > p.config.MaxRPS {
				currentRPS = p.config.MaxRPS
			}
			p.rateLimiter.UpdateRate(currentRPS)
		}
	}
}

// worker processes requests from the jobs channel
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case _, ok := <-p.jobs:
			if !ok {
				return
			}
			if err := p.rateLimiter.Wait(ctx); err != nil {
				return
			}
			p.executeRequest(ctx)
		}
	}
}

// executeRequest performs a single request and updates metrics
func (p *Pool) executeRequest(ctx context.Context) {
	start := time.Now()

	// Select a random endpoint
	endpoint := p.config.Endpoints[rand.Intn(len(p.config.Endpoints))]

	// Create request
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// Set method and URL
	req.Header.SetMethod(endpoint.Method)
	req.SetRequestURI(endpoint.URL)

	// Process and set headers
	headers, err := p.processor.ProcessMap(endpoint.Headers)
	if err != nil {
		p.updateMetrics(start, false)
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Process and set query parameters
	if len(endpoint.QueryParams) > 0 {
		queryParams, err := p.processor.ProcessMap(endpoint.QueryParams)
		if err != nil {
			p.updateMetrics(start, false)
			return
		}
		q := req.URI().QueryArgs()
		for k, v := range queryParams {
			q.Add(k, v)
		}
	}

	// Process and set body
	if endpoint.Body != nil {
		bodyBytes, err := json.Marshal(endpoint.Body)
		if err != nil {
			p.updateMetrics(start, false)
			return
		}
		bodyStr := string(bodyBytes)
		processedBody, err := p.processor.ProcessTemplate(bodyStr, nil)
		if err != nil {
			p.updateMetrics(start, false)
			return
		}
		req.SetBodyString(processedBody)
	}

	// Execute request
	err = p.client.Do(req, resp)
	success := err == nil && resp.StatusCode() >= 200 && resp.StatusCode() < 300

	p.updateMetrics(start, success)
}

// updateMetrics updates the metrics with the request results
func (p *Pool) updateMetrics(start time.Time, success bool) {
	duration := time.Since(start)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.TotalRequests++
	if success {
		p.metrics.SuccessfulRequests++
	} else {
		p.metrics.FailedRequests++
	}

	// Update latency stats
	p.latencies = append(p.latencies, duration)
	if len(p.latencies) > 1000 {
		p.latencies = p.latencies[1:]
	}

	// Update min/max
	if p.metrics.LatencyStats.Min == 0 || duration < p.metrics.LatencyStats.Min {
		p.metrics.LatencyStats.Min = duration
	}
	if duration > p.metrics.LatencyStats.Max {
		p.metrics.LatencyStats.Max = duration
	}

	// Calculate mean
	var total time.Duration
	for _, lat := range p.latencies {
		total += lat
	}
	p.metrics.LatencyStats.Mean = total / time.Duration(len(p.latencies))

	// Calculate percentiles
	if len(p.latencies) > 0 {
		p.metrics.LatencyStats.P50 = p.calculatePercentile(0.5)
		p.metrics.LatencyStats.P95 = p.calculatePercentile(0.95)
		p.metrics.LatencyStats.P99 = p.calculatePercentile(0.99)
	}
}

// calculatePercentile calculates the given percentile from the latency data
func (p *Pool) calculatePercentile(percentile float64) time.Duration {
	if len(p.latencies) == 0 {
		return 0
	}

	index := int(float64(len(p.latencies)-1) * percentile)
	return p.latencies[index]
}

// Stop gracefully shuts down the worker pool
func (p *Pool) Stop() {
	close(p.stopChan) // Signal all goroutines to stop
	p.wg.Wait()       // Wait for all goroutines to finish
	close(p.jobs)     // Close the jobs channel after all workers are done
}

// GetMetrics returns the current metrics
func (p *Pool) GetMetrics() *config.Metrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.metrics
}
