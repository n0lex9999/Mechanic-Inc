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
func Worker(id int, targets <-chan string, results chan<- Result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range targets {
		start := time.Now()
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			results <- Result{URL: url, Err: err}
			continue
		}

		resp, err := client.Do(req)
		duration := time.Since(start)

		res := Result{
			URL:      url,
			Duration: duration,
			Err:      err,
		}

		if err == nil {
			res.StatusCode = resp.StatusCode
			// Effleurement du body pour permettre la réutilisation du socket TCP
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		results <- res
	}
}
