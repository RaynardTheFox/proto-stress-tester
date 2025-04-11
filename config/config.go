package config

import (
	"time"
)

// Config represents the main configuration structure
type Config struct {
	Endpoints   []Endpoint    `yaml:"endpoints"`
	LoadPattern LoadPattern   `yaml:"load_pattern"`
	Duration    time.Duration `yaml:"duration"`
	MaxRPS      int           `yaml:"max_rps"`
}

// Endpoint represents a single API endpoint configuration
type Endpoint struct {
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`
	Headers     map[string]string `yaml:"headers"`
	QueryParams map[string]string `yaml:"query_params"`
	Body        interface{}       `yaml:"body"`
}

// LoadPattern defines how the load should be applied
type LoadPattern struct {
	Type      string        `yaml:"type"` // constant, ramp-up, spike
	StartRPS  int           `yaml:"start_rps"`
	Increment int           `yaml:"increment"`
	Interval  time.Duration `yaml:"interval"`
}

// Metrics represents the collected metrics during the test
type Metrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	LatencyStats       LatencyStats
	CurrentRPS         float64
}

// LatencyStats contains latency distribution statistics
type LatencyStats struct {
	Min  time.Duration
	Max  time.Duration
	Mean time.Duration
	P50  time.Duration
	P95  time.Duration
	P99  time.Duration
}
