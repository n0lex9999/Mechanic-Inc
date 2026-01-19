package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}

func StartWebServer(port int) {
	mux := http.NewServeMux()

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			fmt.Printf("[LOG] %s %s | %v | %s\n", r.Method, r.URL.Path, time.Since(start), r.RemoteAddr)
		})
	}

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/probe", handleProbe)
	mux.HandleFunc("/api/probe-stream", handleProbeStream) // NEW: Streaming endpoint

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		loggingMiddleware(mux).ServeHTTP(w, r)
	})

	fmt.Printf("[*] MECHANIC CORE - Engine started on port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Printf("[CRITICAL] Server failure: %v\n", err)
	}
}

type ProbeRequest struct {
	URL         string `json:"url"`
	Concurrency int    `json:"concurrency"`
	Count       int    `json:"count"`
	Timeout     int    `json:"timeout"`
}

// handleProbe - Original endpoint for small requests
func handleProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	sanitizeRequest(&req)

	fmt.Printf("[EXE] Starting Probe -> Target: %s | Workers: %d | Total: %d\n", req.URL, req.Concurrency, req.Count)

	start := time.Now()
	stats := PerformProbe(req.URL, req.Concurrency, req.Count, req.Timeout)
	duration := time.Since(start)

	fmt.Printf("[RES] Probe Finished -> Success: %d | Errors: %d | Time: %v\n", stats.SuccessCount, stats.ErrorCount, duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"results":    stats,
		"latency_ms": stats.AvgLatency.Milliseconds(),
		"total_time": duration.String(),
	})
}

// handleProbeStream - STREAMING endpoint for big requests (prevents 504)
func handleProbeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	sanitizeRequest(&req)

	// Setup streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[STREAM] Starting -> Target: %s | Workers: %d | Total: %d\n", req.URL, req.Concurrency, req.Count)

	// Setup probe
	if req.Timeout > 5 {
		req.Timeout = 5
	}
	client := NewClient(time.Duration(req.Timeout) * time.Second)
	targets := make(chan string, req.Count)
	results := make(chan Result, req.Count)

	var wg sync.WaitGroup
	for i := 0; i < req.Concurrency; i++ {
		wg.Add(1)
		go Worker(i, targets, results, client, &wg)
	}

	for i := 0; i < req.Count; i++ {
		targets <- req.URL
	}
	close(targets)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Stream results in real-time
	successCount := 0
	errorCount := 0
	var totalLatency time.Duration
	processed := 0

	for res := range results {
		processed++
		if res.Err != nil {
			errorCount++
		} else {
			successCount++
			totalLatency += res.Duration
		}

		// Send progress every 50 results (or at the end)
		if processed%50 == 0 || processed == req.Count {
			avgLat := int64(0)
			if successCount > 0 {
				avgLat = (totalLatency / time.Duration(successCount)).Milliseconds()
			}

			data := map[string]interface{}{
				"progress":   processed,
				"total":      req.Count,
				"success":    successCount,
				"errors":     errorCount,
				"latency_ms": avgLat,
			}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}

	// Final message
	avgLat := int64(0)
	if successCount > 0 {
		avgLat = (totalLatency / time.Duration(successCount)).Milliseconds()
	}
	finalData := map[string]interface{}{
		"done":       true,
		"success":    successCount,
		"errors":     errorCount,
		"latency_ms": avgLat,
	}
	jsonData, _ := json.Marshal(finalData)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()

	fmt.Printf("[STREAM] Done -> Success: %d | Errors: %d\n", successCount, errorCount)
}

func sanitizeRequest(req *ProbeRequest) {
	if req.Concurrency <= 0 {
		req.Concurrency = 10
	}
	if req.Count <= 0 {
		req.Count = 100
	}
	if req.Timeout <= 0 {
		req.Timeout = 5
	}
	if req.Concurrency > 1000 {
		req.Concurrency = 1000
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "http://" + req.URL
	}
}
