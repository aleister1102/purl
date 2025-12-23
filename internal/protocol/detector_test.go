package protocol

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/aleister1102/purl/internal/cli"
	"github.com/aleister1102/purl/internal/target"
)

// Property 5: Protocol Override Behavior
// For any --proto flag value ("http" or "https"), THE Protocol_Detector SHALL use exactly that protocol
// without attempting the other, regardless of target format.
// Validates: Requirements 2.5, 2.6
func TestProtocolOverrideBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Protocol override uses specified protocol without fallback",
		prop.ForAll(
			func(proto string) bool {
				// Only test valid protocol values
				if proto != "http" && proto != "https" {
					return true // Skip invalid values
				}

				// Start a test server that responds to both HTTP and HTTPS
				httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				}))
				defer httpServer.Close()

				// Extract host and port from the test server
				u, _ := url.Parse(httpServer.URL)
				host := u.Hostname()
				port := u.Port()

				// Create a parsed target
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   net.JoinHostPort(host, port),
						Path:   "/",
					},
					IsIP:             false,
					HasExplicitProto: false,
				}

				// Create options with the specified protocol
				opts := &cli.Options{
					Proto:          proto,
					Timeout:        5 * time.Second,
					ConnectTimeout: 5 * time.Second,
				}

				// Call DetectProtocol
				result, _ := DetectProtocol(parsedTarget, opts)

				// Verify the protocol matches what was specified
				if result == nil {
					return false
				}

				// The protocol in the result should match the specified protocol
				// (or be the one we attempted if it's the only one available)
				if proto == "http" {
					// When proto is "http", we should attempt HTTP
					// The result protocol should be "http" if successful
					if result.Protocol != "http" {
						return false
					}
				} else if proto == "https" {
					// When proto is "https", we should attempt HTTPS
					// The result protocol should be "https"
					if result.Protocol != "https" {
						return false
					}
				}

				return true
			},
			gen.OneConstOf("http", "https"),
		),
	)

	if !properties.Run(gopter.NewFormatedReporter(true, 160, io.Discard)) {
		t.Fail()
	}
}

// Test that auto mode tries HTTP first
func TestAutoModeTriesHTTPFirst(t *testing.T) {
	// Start an HTTP server that responds
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer httpServer.Close()

	u, _ := url.Parse(httpServer.URL)
	host := u.Hostname()
	port := u.Port()

	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	opts := &cli.Options{
		Proto:          "auto",
		Timeout:        5 * time.Second,
		ConnectTimeout: 5 * time.Second,
	}

	result, _ := DetectProtocol(parsedTarget, opts)

	if result.Protocol != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", result.Protocol)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", result.StatusCode)
	}
}

// Test that manual protocol mode doesn't attempt fallback
func TestManualProtocolNoFallback(t *testing.T) {
	// Create a target that will fail on HTTPS but succeed on HTTP
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer httpServer.Close()

	u, _ := url.Parse(httpServer.URL)
	host := u.Hostname()
	port := u.Port()

	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	// Specify HTTPS manually (which will fail since we only have HTTP server)
	opts := &cli.Options{
		Proto:          "https",
		Timeout:        2 * time.Second,
		ConnectTimeout: 2 * time.Second,
	}

	result, _ := DetectProtocol(parsedTarget, opts)

	// Should attempt HTTPS and fail, not fall back to HTTP
	if result.Protocol != "https" {
		t.Errorf("Expected protocol 'https', got '%s'", result.Protocol)
	}

	// Should have an error since HTTPS will fail
	if result.Error == nil {
		t.Error("Expected error when connecting to HTTP server with HTTPS protocol")
	}
}

// Test ProbeResult structure
func TestProbeResultStructure(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer httpServer.Close()

	u, _ := url.Parse(httpServer.URL)
	host := u.Hostname()
	port := u.Port()

	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	opts := &cli.Options{
		Proto:          "http",
		Timeout:        5 * time.Second,
		ConnectTimeout: 5 * time.Second,
	}

	result, err := DetectProtocol(parsedTarget, opts)

	if err != nil {
		t.Fatalf("DetectProtocol failed: %v", err)
	}

	// Verify ProbeResult fields
	if result.Protocol == "" {
		t.Error("Protocol should not be empty")
	}

	if result.StatusCode == 0 {
		t.Error("StatusCode should be set")
	}

	if result.Duration == 0 {
		t.Error("Duration should be measured")
	}

	if result.Response == nil {
		t.Error("Response should be set on success")
	}
}

// Test that duration is measured
func TestDurationMeasurement(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some delay
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer httpServer.Close()

	u, _ := url.Parse(httpServer.URL)
	host := u.Hostname()
	port := u.Port()

	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	opts := &cli.Options{
		Proto:          "http",
		Timeout:        5 * time.Second,
		ConnectTimeout: 5 * time.Second,
	}

	result, err := DetectProtocol(parsedTarget, opts)

	if err != nil {
		t.Fatalf("DetectProtocol failed: %v", err)
	}

	// Duration should be at least 100ms (the sleep time)
	if result.Duration < 100*time.Millisecond {
		t.Errorf("Expected duration >= 100ms, got %v", result.Duration)
	}
}

// Test timeout behavior
func TestTimeoutBehavior(t *testing.T) {
	// Create a server that never responds
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	host, port, _ := net.SplitHostPort(addr)

	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	opts := &cli.Options{
		Proto:          "http",
		Timeout:        500 * time.Millisecond,
		ConnectTimeout: 500 * time.Millisecond,
	}

	startTime := time.Now()
	result, _ := DetectProtocol(parsedTarget, opts)
	elapsed := time.Since(startTime)

	// Should timeout around 500ms
	if elapsed < 400*time.Millisecond {
		t.Errorf("Expected timeout around 500ms, got %v", elapsed)
	}

	if result.Error == nil {
		t.Error("Expected timeout error")
	}
}

// Test status code capture
func TestStatusCodeCapture(t *testing.T) {
	testCases := []int{200, 301, 404, 500}

	for _, statusCode := range testCases {
		t.Run(fmt.Sprintf("StatusCode_%d", statusCode), func(t *testing.T) {
			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
				w.Write([]byte("OK"))
			}))
			defer httpServer.Close()

			u, _ := url.Parse(httpServer.URL)
			host := u.Hostname()
			port := u.Port()

			parsedTarget := &target.ParsedTarget{
				URL: &url.URL{
					Scheme: "http",
					Host:   net.JoinHostPort(host, port),
					Path:   "/",
				},
				IsIP:             false,
				HasExplicitProto: false,
			}

			opts := &cli.Options{
				Proto:          "http",
				Timeout:        5 * time.Second,
				ConnectTimeout: 5 * time.Second,
			}

			result, err := DetectProtocol(parsedTarget, opts)

			if err != nil {
				t.Fatalf("DetectProtocol failed: %v", err)
			}

			if result.StatusCode != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, result.StatusCode)
			}
		})
	}
}
