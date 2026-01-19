package main

import (
	"runtime"
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
	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Shorter timeout for faster failure detection
	if timeoutSec > 5 {
		timeoutSec = 5
	}

	client := NewClient(time.Duration(timeoutSec) * time.Second)

	// Buffered channels for zero-blocking
	targets := make(chan string, requestCount)
	results := make(chan Result, requestCount)

	var wg sync.WaitGroup

	// Start workers BEFORE feeding targets (pipeline optimization)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go Worker(i, targets, results, client, &wg)
	}

	// Feed all targets instantly (non-blocking because buffer is big enough)
	for i := 0; i < requestCount; i++ {
		targets <- targetURL
	}
	close(targets)

	// Collect results in background
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
