package protocol

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/user/curlz/internal/cli"
	"github.com/user/curlz/internal/errors"
	"github.com/user/curlz/internal/target"
	"github.com/user/curlz/internal/transport"
)

// ProbeResult contains the result of protocol detection
type ProbeResult struct {
	Protocol   string // "http" or "https"
	StatusCode int
	Duration   time.Duration
	Response   *http.Response
	Error      error
}

// DetectProtocol probes the target and returns the working protocol
// In auto mode: tries HTTP first (3s timeout), then HTTPS (7s timeout)
// In manual mode: uses the specified protocol directly
func DetectProtocol(parsedTarget *target.ParsedTarget, opts *cli.Options) (*ProbeResult, error) {
	// If protocol is manually specified, use it directly
	if opts.Proto != "" && opts.Proto != "auto" {
		result := probeProtocol(parsedTarget, opts, opts.Proto)
		return result, result.Error
	}

	// Auto mode: try HTTP first, then HTTPS
	// Try HTTP with 3 second timeout
	httpResult := probeProtocolWithTimeout(parsedTarget, opts, "http", 3*time.Second)
	if httpResult.Error == nil && httpResult.StatusCode >= 200 && httpResult.StatusCode < 400 {
		// HTTP succeeded with success status
		return httpResult, nil
	}

	// HTTP failed or returned non-success status, try HTTPS with 7 second timeout
	httpsResult := probeProtocolWithTimeout(parsedTarget, opts, "https", 7*time.Second)
	if httpsResult.Error == nil {
		// HTTPS succeeded
		return httpsResult, nil
	}

	// Both failed, return the HTTPS result with error
	return httpsResult, nil
}

// probeProtocol attempts to connect using the specified protocol
// Uses the default timeout from opts
func probeProtocol(parsedTarget *target.ParsedTarget, opts *cli.Options, proto string) *ProbeResult {
	timeout := transport.ApplyTimeouts(opts)
	return probeProtocolWithTimeout(parsedTarget, opts, proto, timeout)
}

// probeProtocolWithTimeout attempts to connect using the specified protocol and timeout
func probeProtocolWithTimeout(parsedTarget *target.ParsedTarget, opts *cli.Options, proto string, timeout time.Duration) *ProbeResult {
	result := &ProbeResult{
		Protocol: proto,
	}

	// Create a copy of options with the specified protocol
	probeOpts := *opts
	probeOpts.Timeout = timeout
	probeOpts.ConnectTimeout = timeout

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create transport for this protocol
	tr, err := transport.NewTransport(&probeOpts, parsedTarget)
	if err != nil {
		result.Error = err
		return result
	}

	// Create HTTP client with the transport
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	// Construct the URL with the specified protocol
	probeURL := constructURL(parsedTarget, proto)

	// Create a HEAD request (lightweight probe)
	req, err := http.NewRequestWithContext(ctx, "HEAD", probeURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// Record start time
	startTime := time.Now()

	// Execute the request
	resp, err := client.Do(req)
	duration := time.Since(startTime)
	result.Duration = duration

	if err != nil {
		result.Error = mapProbeError(err, parsedTarget)
		return result
	}

	// Store response and status code
	result.Response = resp
	result.StatusCode = resp.StatusCode

	return result
}

// constructURL constructs a URL with the specified protocol
func constructURL(parsedTarget *target.ParsedTarget, proto string) string {
	// Create a new URL with the specified scheme
	probeURL := *parsedTarget.URL
	probeURL.Scheme = proto
	return probeURL.String()
}

// mapProbeError maps network errors to appropriate error types
func mapProbeError(err error, parsedTarget *target.ParsedTarget) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for timeout
	if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return &errors.TimeoutError{
			Duration: 0, // Duration is tracked separately in ProbeResult
			Phase:    "request",
		}
	}

	// Check for connection refused or no route
	if parsedTarget != nil && parsedTarget.URL != nil {
		host := parsedTarget.URL.Hostname()
		if host != "" {
			// Check for common error patterns
			if contains(errStr, "connection refused") {
				return &errors.ConnectionError{
					Host:  host,
					Port:  parsedTarget.URL.Port(),
					Cause: err,
				}
			}
			if contains(errStr, "no such host") || contains(errStr, "name resolution") {
				return &errors.NoRouteError{
					Host:  host,
					Cause: err,
				}
			}
			if contains(errStr, "certificate") || contains(errStr, "tls") || contains(errStr, "ssl") {
				return &errors.TLSError{
					Host:  host,
					Cause: err,
				}
			}
		}
	}

	// Default to connection error
	return &errors.ConnectionError{
		Host:  "unknown",
		Port:  "unknown",
		Cause: err,
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0))
}
