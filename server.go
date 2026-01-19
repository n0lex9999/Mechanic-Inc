package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}

func StartWebServer(port int) {
	mux := http.NewServeMux()

	// Better logging middleware
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			fmt.Printf("[LOG] %s %s | %v | %s\n", r.Method, r.URL.Path, time.Since(start), r.RemoteAddr)
		})
	}

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/probe", handleProbe)

	// Robust CORS
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

func handleProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("[ERR] Invalid JSON payload from %s\n", r.RemoteAddr)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Sanitize & Default
	if req.Concurrency <= 0 {
		req.Concurrency = 10
	}
	if req.Count <= 0 {
		req.Count = 100
	}
	if req.Timeout <= 0 {
		req.Timeout = 10
	}
	if req.Concurrency > 1000 {
		req.Concurrency = 1000
	} // High-performance limit

	// IP Support: if it doesn't start with http, assume it's a raw IP/domain
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "http://" + req.URL
		fmt.Printf("[INF] Raw address detected, formatted to: %s\n", req.URL)
	}

	fmt.Printf("[EXE] Starting Probe -> Target: %s | Workers: %d | Total: %d\n", req.URL, req.Concurrency, req.Count)

	// Core execution with performance monitoring
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
