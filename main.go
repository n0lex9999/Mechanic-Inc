package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	targetURL := flag.String("u", "", "Target URL (e.g., http://example.com)")
	concurrency := flag.Int("c", 10, "Concurrency level (number of workers)")
	requestCount := flag.Int("n", 100, "Number of requests to send")
	timeoutSec := flag.Int("t", 5, "Request timeout in seconds")
	webMode := flag.Bool("web", false, "Start web dashboard")
	port := flag.Int("port", 8080, "Port for web dashboard")

	flag.Parse()

	if *webMode {
		StartWebServer(*port)
		return
	}

	if *targetURL == "" {
		fmt.Println("Usage: goprobe -u <target_url> [-c concurrency] [-n count] [-t timeout] [-web]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("[*] Starting %d workers for target: %s\n", *concurrency, *targetURL)
	stats := PerformProbe(*targetURL, *concurrency, *requestCount, *timeoutSec)

	fmt.Printf("\n--- Statistics for %s ---\n", stats.TargetURL)
	fmt.Printf("Total Requests: %d\n", stats.TotalRequest)
	fmt.Printf("Successful:     %d\n", stats.SuccessCount)
	fmt.Printf("Failed:         %d\n", stats.ErrorCount)
	if stats.SuccessCount > 0 {
		fmt.Printf("Avg Latency:    %v\n", stats.AvgLatency)
	}
}
