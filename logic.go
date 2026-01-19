package main

import (
	"sync"
	"time"
)

type ProbeStats struct {
	TargetURL    string        `json:"target_url"`
	TotalRequest int           `json:"total_requests"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

func PerformProbe(targetURL string, concurrency int, requestCount int, timeoutSec int) ProbeStats {
	client := NewClient(time.Duration(timeoutSec) * time.Second)
	targets := make(chan string, requestCount)
	results := make(chan Result, requestCount)

	var wg sync.WaitGroup

	// Start workers
	for i := 1; i <= concurrency; i++ {
		wg.Add(1)
		go Worker(i, targets, results, client, &wg)
	}

	// Feed targets
	go func() {
		for i := 0; i < requestCount; i++ {
			targets <- targetURL
		}
		close(targets)
	}()

	// Wait and close
	go func() {
		wg.Wait()
		close(results)
	}()

	stats := ProbeStats{
		TargetURL:    targetURL,
		TotalRequest: requestCount,
	}

	var totalTime time.Duration
	for res := range results {
		if res.Err != nil {
			stats.ErrorCount++
		} else {
			stats.SuccessCount++
			totalTime += res.Duration
		}
	}

	if stats.SuccessCount > 0 {
		stats.AvgLatency = totalTime / time.Duration(stats.SuccessCount)
	}

	return stats
}
