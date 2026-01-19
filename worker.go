package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// Result holds the outcome of a request
type Result struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	Err        error
}

// Worker process targets from a channel and sends results back
// OPTIMIZED: Minimal allocations, fast body drain
func Worker(id int, targets <-chan string, results chan<- Result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	// Reusable buffer for draining body
	buf := make([]byte, 512)

	for url := range targets {
		start := time.Now()

		req, err := http.NewRequest("HEAD", url, nil) // HEAD is faster than GET
		if err != nil {
			results <- Result{URL: url, Err: err}
			continue
		}

		// Minimal headers for speed
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		duration := time.Since(start)

		res := Result{
			URL:      url,
			Duration: duration,
			Err:      err,
		}

		if err == nil {
			res.StatusCode = resp.StatusCode
			// Fast body drain using small buffer
			io.CopyBuffer(io.Discard, resp.Body, buf)
			resp.Body.Close()
		}

		results <- res
	}
}
