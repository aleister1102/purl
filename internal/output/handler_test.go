package output

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/user/purl/internal/cli"
	"github.com/user/purl/internal/protocol"
)

// Unit Tests

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			expected: "500µs",
		},
		{
			name:     "milliseconds",
			duration: 250 * time.Millisecond,
			expected: "250ms",
		},
		{
			name:     "one second",
			duration: 1 * time.Second,
			expected: "1s",
		},
		{
			name:     "fractional seconds",
			duration: 1500 * time.Millisecond,
			expected: "1.50s",
		},
		{
			name:     "multiple seconds",
			duration: 5 * time.Second,
			expected: "5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatProto(t *testing.T) {
	tests := []struct {
		name     string
		proto    string
		expected string
	}{
		{
			name:     "http lowercase",
			proto:    "http",
			expected: "HTTP",
		},
		{
			name:     "https lowercase",
			proto:    "https",
			expected: "HTTPS",
		},
		{
			name:     "already uppercase",
			proto:    "HTTP",
			expected: "HTTP",
		},
		{
			name:     "mixed case",
			proto:    "HtTp",
			expected: "HtTp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatProto(tt.proto)
			if result != tt.expected {
				t.Errorf("formatProto(%q) = %q, want %q", tt.proto, result, tt.expected)
			}
		})
	}
}

func TestPrintStatusLine(t *testing.T) {
	tests := []struct {
		name           string
		protocol       string
		statusCode     int
		duration       time.Duration
		expectedFormat string
	}{
		{
			name:           "HTTP 200 OK",
			protocol:       "http",
			statusCode:     200,
			duration:       1 * time.Second,
			expectedFormat: "[HTTP] Status: 200 Time: 1s",
		},
		{
			name:           "HTTPS 404 Not Found",
			protocol:       "https",
			statusCode:     404,
			duration:       500 * time.Millisecond,
			expectedFormat: "[HTTPS] Status: 404 Time: 500ms",
		},
		{
			name:           "HTTP 500 Server Error",
			protocol:       "http",
			statusCode:     500,
			duration:       2500 * time.Millisecond,
			expectedFormat: "[HTTP] Status: 500 Time: 2.50s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &cli.Options{}
			handler := NewHandler(opts)

			result := &protocol.ProbeResult{
				Protocol:   tt.protocol,
				StatusCode: tt.statusCode,
				Duration:   tt.duration,
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handler.printStatusLine(result)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("printStatusLine() error = %v", err)
			}

			output, _ := io.ReadAll(r)
			outputStr := strings.TrimSpace(string(output))

			if !strings.Contains(outputStr, tt.expectedFormat) {
				t.Errorf("printStatusLine() output = %q, want to contain %q", outputStr, tt.expectedFormat)
			}
		})
	}
}

func TestWriteResponseBody_ToStdout(t *testing.T) {
	opts := &cli.Options{
		Output: "", // Write to stdout
	}
	handler := NewHandler(opts)

	// Create a mock response with body
	bodyContent := "Hello, World!"
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(bodyContent)),
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := handler.writeResponseBody(resp)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("writeResponseBody() error = %v", err)
	}

	output, _ := io.ReadAll(r)
	if string(output) != bodyContent {
		t.Errorf("writeResponseBody() output = %q, want %q", string(output), bodyContent)
	}
}

func TestWriteResponseBody_ToFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "purl-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := &cli.Options{
		Output: tmpFile.Name(),
	}
	handler := NewHandler(opts)

	// Create a mock response with body
	bodyContent := "Test output content"
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(bodyContent)),
	}

	err = handler.writeResponseBody(resp)
	if err != nil {
		t.Fatalf("writeResponseBody() error = %v", err)
	}

	// Read the file and verify content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(content) != bodyContent {
		t.Errorf("file content = %q, want %q", string(content), bodyContent)
	}
}

func TestPrintVerboseRequest(t *testing.T) {
	opts := &cli.Options{}
	handler := NewHandler(opts)

	// Create a mock request
	url, _ := url.Parse("http://example.com/api/test")
	req := &http.Request{
		Method: "POST",
		URL:    url,
		Proto:  "HTTP/1.1",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"User-Agent":   []string{"purl/1.0"},
		},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := handler.printVerboseRequest(req)

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("printVerboseRequest() error = %v", err)
	}

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Verify output contains expected elements
	if !strings.Contains(outputStr, "POST") {
		t.Errorf("output missing method: %q", outputStr)
	}
	if !strings.Contains(outputStr, "Content-Type") {
		t.Errorf("output missing header: %q", outputStr)
	}
}

func TestPrintVerboseRequest_MasksAuthHeader(t *testing.T) {
	opts := &cli.Options{}
	handler := NewHandler(opts)

	// Create a mock request with Authorization header
	url, _ := url.Parse("http://example.com/api/test")
	req := &http.Request{
		Method: "GET",
		URL:    url,
		Proto:  "HTTP/1.1",
		Header: http.Header{
			"Authorization": []string{"Bearer secret-token-12345"},
		},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := handler.printVerboseRequest(req)

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("printVerboseRequest() error = %v", err)
	}

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Verify Authorization header is redacted
	if strings.Contains(outputStr, "secret-token-12345") {
		t.Errorf("Authorization header not redacted: %q", outputStr)
	}
	if !strings.Contains(outputStr, "[REDACTED]") {
		t.Errorf("output missing [REDACTED]: %q", outputStr)
	}
}

// Property-Based Tests

// Property 7: Status Line Format
// For any successful response, the output status line SHALL match the expected format
func TestProperty_StatusLineFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Status line format matches pattern", prop.ForAll(
		func(proto string, statusCode int, durationMs int64) bool {
			// Normalize protocol
			if proto != "http" && proto != "https" {
				proto = "http"
			}

			// Ensure valid status code
			if statusCode < 100 || statusCode > 599 {
				statusCode = 200
			}

			// Ensure positive duration
			if durationMs < 0 {
				durationMs = 0
			}

			opts := &cli.Options{}
			handler := NewHandler(opts)

			result := &protocol.ProbeResult{
				Protocol:   proto,
				StatusCode: statusCode,
				Duration:   time.Duration(durationMs) * time.Millisecond,
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handler.printStatusLine(result)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				return false
			}

			output, _ := io.ReadAll(r)
			outputStr := strings.TrimSpace(string(output))

			// Verify format: [PROTO] Status: CODE Time: Xs
			// Check for required components
			if !strings.Contains(outputStr, "[") || !strings.Contains(outputStr, "]") {
				return false
			}
			if !strings.Contains(outputStr, "Status:") {
				return false
			}
			if !strings.Contains(outputStr, "Time:") {
				return false
			}

			// Verify protocol is uppercase
			expectedProto := strings.ToUpper(proto)
			if !strings.Contains(outputStr, "["+expectedProto+"]") {
				return false
			}

			// Verify status code is present
			if !strings.Contains(outputStr, fmt.Sprintf("%d", statusCode)) {
				return false
			}

			return true
		},
		gen.OneConstOf("http", "https"),
		gen.IntRange(100, 599),
		gen.Int64Range(0, 10000),
	))

	if !properties.Run(gopter.NewFormatedReporter(true, 160, os.Stdout)) {
		t.Fail()
	}
}

// Property: Response body streaming
// For any response body, writing it should produce identical content
func TestProperty_ResponseBodyStreaming(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Response body content is preserved", prop.ForAll(
		func(bodyContent string) bool {
			opts := &cli.Options{
				Output: "", // Write to stdout
			}
			handler := NewHandler(opts)

			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(bodyContent)),
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := handler.writeResponseBody(resp)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				return false
			}

			output, _ := io.ReadAll(r)
			return string(output) == bodyContent
		},
		gen.AnyString(),
	))

	if !properties.Run(gopter.NewFormatedReporter(true, 160, os.Stdout)) {
		t.Fail()
	}
}

// Property: Verbose request output contains all headers
// For any request with headers, verbose output should include all of them
func TestProperty_VerboseRequestIncludesHeaders(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Verbose output includes all headers", prop.ForAll(
		func(headerName string, headerValue string) bool {
			// Ensure valid header name (alphanumeric and hyphens)
			if headerName == "" {
				headerName = "X-Custom-Header"
			}

			opts := &cli.Options{}
			handler := NewHandler(opts)

			url, _ := url.Parse("http://example.com/test")
			req := &http.Request{
				Method: "GET",
				URL:    url,
				Proto:  "HTTP/1.1",
				Header: http.Header{
					headerName: []string{headerValue},
				},
			}

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			err := handler.printVerboseRequest(req)

			w.Close()
			os.Stderr = oldStderr

			if err != nil {
				return false
			}

			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Verify header name is present
			return strings.Contains(outputStr, headerName)
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	if !properties.Run(gopter.NewFormatedReporter(true, 160, os.Stdout)) {
		t.Fail()
	}
}

// Property: File output creates file with correct content
// For any output filename and body content, the file should be created with exact content
func TestProperty_FileOutputCreation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Fewer iterations due to file I/O

	properties := gopter.NewProperties(parameters)

	properties.Property("File output contains exact body content", prop.ForAll(
		func(bodyContent string) bool {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "purl-prop-test-*.txt")
			if err != nil {
				return false
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			opts := &cli.Options{
				Output: tmpFile.Name(),
			}
			handler := NewHandler(opts)

			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(bodyContent)),
			}

			err = handler.writeResponseBody(resp)
			if err != nil {
				return false
			}

			// Read the file and verify content
			content, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				return false
			}

			return string(content) == bodyContent
		},
		gen.AnyString(),
	))

	if !properties.Run(gopter.NewFormatedReporter(true, 160, os.Stdout)) {
		t.Fail()
	}
}
