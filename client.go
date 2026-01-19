package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewClient returns an ULTRA-OPTIMIZED http.Client for maximum throughput.
func NewClient(timeout time.Duration) *http.Client {
	// Custom dialer with aggressive timeouts
	dialer := &net.Dialer{
		Timeout:   3 * time.Second,  // Fast connection timeout
		KeepAlive: 60 * time.Second, // Keep connections alive
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true, // Use HTTP/2 for speed
		MaxIdleConns:          1000, // HUGE connection pool
		MaxIdleConnsPerHost:   500,  // Many connections per target
		MaxConnsPerHost:       0,    // Unlimited
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second, // Fast TLS
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true, // Skip decompression overhead
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip cert validation
			MinVersion:         tls.VersionTLS12,
		},
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - saves time
			return http.ErrUseLastResponse
		},
	}
}
