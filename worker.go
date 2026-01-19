package main

import (
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
func Worker(id int, targets <-chan string, results chan<- Result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range targets {
		start := time.Now()
		resp, err := client.Get(url)
		duration := time.Since(start)

		res := Result{
			URL:      url,
			Duration: duration,
			Err:      err,
		}

		if err == nil {
			res.StatusCode = resp.StatusCode
			resp.Body.Close()
		}

		results <- res
	}
}
