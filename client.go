package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewClient returns an optimized http.Client for high-reliability network requests.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// Ignore TLS certificate verification for TP environments
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
