package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func StartWebServer(port int) {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/probe", handleProbe)

	fmt.Printf("[*] Web Dashboard started on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("[!] Server error: %v\n", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// We'll serve the HTML content here (defined in another file or as a string)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, dashboardHTML)
}

type ProbeRequest struct {
	URL         string `json:"url"`
	Concurrency int    `json:"concurrency"`
	Count       int    `json:"count"`
	Timeout     int    `json:"timeout"`
}

func handleProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default values if not provided
	if req.Concurrency == 0 {
		req.Concurrency = 10
	}
	if req.Count == 0 {
		req.Count = 100
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}

	stats := PerformProbe(req.URL, req.Concurrency, req.Count, req.Timeout)

	w.Header().Set("Content-Type", "json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"results":    stats,
		"latency_ms": stats.AvgLatency.Milliseconds(),
	})
}
