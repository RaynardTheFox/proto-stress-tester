package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"protobuf/config"
	"protobuf/worker"

	"github.com/spf13/viper"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	duration := flag.Duration("duration", 5*time.Minute, "Test duration")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Error: Configuration file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line flags
	cfg.Duration = *duration

	// Create worker pool
	pool := worker.NewPool(*workers, cfg)

	// Setup context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Duration)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Start the stress test
	fmt.Printf("Starting stress test with %d workers for %v\n", *workers, cfg.Duration)
	pool.Start(ctx)

	// Wait for completion or interruption
	<-ctx.Done()

	// Stop the pool and print results
	pool.Stop()
	printResults(pool.GetMetrics())
}

func loadConfig(configFile string) (*config.Config, error) {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func printResults(metrics *config.Metrics) {
	fmt.Println("\nTest Results:")
	fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", metrics.FailedRequests)
	fmt.Printf("Current RPS: %.2f\n", metrics.CurrentRPS)
	fmt.Println("\nLatency Statistics:")
	fmt.Printf("Min: %v\n", metrics.LatencyStats.Min)
	fmt.Printf("Max: %v\n", metrics.LatencyStats.Max)
	fmt.Printf("Mean: %v\n", metrics.LatencyStats.Mean)
	fmt.Printf("P50: %v\n", metrics.LatencyStats.P50)
	fmt.Printf("P95: %v\n", metrics.LatencyStats.P95)
	fmt.Printf("P99: %v\n", metrics.LatencyStats.P99)
}
