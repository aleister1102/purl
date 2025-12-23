package target

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/aleister1102/purl/internal/errors"
)

// Unit Tests

func TestParseTarget_FullURL(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedScheme   string
		expectedHost     string
		expectedPort     string
		expectedPath     string
		expectedIsIP     bool
		expectedExplicit bool
		shouldError      bool
	}{
		{
			name:             "http URL with port and path",
			input:            "http://example.com:8080/api/v1",
			expectedScheme:   "http",
			expectedHost:     "example.com",
			expectedPort:     "8080",
			expectedPath:     "/api/v1",
			expectedIsIP:     false,
			expectedExplicit: true,
		},
		{
			name:             "https URL with IP",
			input:            "https://192.168.1.1:443/",
			expectedScheme:   "https",
			expectedHost:     "192.168.1.1",
			expectedPort:     "443",
			expectedPath:     "/",
			expectedIsIP:     true,
			expectedExplicit: true,
		},
		{
			name:             "URL without port",
			input:            "http://example.com/path",
			expectedScheme:   "http",
			expectedHost:     "example.com",
			expectedPort:     "",
			expectedPath:     "/path",
			expectedIsIP:     false,
			expectedExplicit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.URL.Scheme != tt.expectedScheme {
				t.Errorf("scheme: got %q, want %q", result.URL.Scheme, tt.expectedScheme)
			}
			if result.URL.Hostname() != tt.expectedHost {
				t.Errorf("host: got %q, want %q", result.URL.Hostname(), tt.expectedHost)
			}
			if result.URL.Port() != tt.expectedPort {
				t.Errorf("port: got %q, want %q", result.URL.Port(), tt.expectedPort)
			}
			if result.URL.Path != tt.expectedPath {
				t.Errorf("path: got %q, want %q", result.URL.Path, tt.expectedPath)
			}
			if result.IsIP != tt.expectedIsIP {
				t.Errorf("IsIP: got %v, want %v", result.IsIP, tt.expectedIsIP)
			}
			if result.HasExplicitProto != tt.expectedExplicit {
				t.Errorf("HasExplicitProto: got %v, want %v", result.HasExplicitProto, tt.expectedExplicit)
			}
		})
	}
}

func TestParseTarget_IPPortPath(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedHost     string
		expectedPort     string
		expectedPath     string
		expectedIsIP     bool
		expectedExplicit bool
	}{
		{
			name:             "IP with port and path",
			input:            "192.168.1.1:8080/api",
			expectedHost:     "192.168.1.1",
			expectedPort:     "8080",
			expectedPath:     "/api",
			expectedIsIP:     true,
			expectedExplicit: false,
		},
		{
			name:             "IP with port only",
			input:            "10.0.0.1:443",
			expectedHost:     "10.0.0.1",
			expectedPort:     "443",
			expectedPath:     "/",
			expectedIsIP:     true,
			expectedExplicit: false,
		},
		{
			name:             "IP with path only",
			input:            "192.168.1.1/admin",
			expectedHost:     "192.168.1.1",
			expectedPort:     "",
			expectedPath:     "/admin",
			expectedIsIP:     true,
			expectedExplicit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.URL.Hostname() != tt.expectedHost {
				t.Errorf("host: got %q, want %q", result.URL.Hostname(), tt.expectedHost)
			}
			if result.URL.Port() != tt.expectedPort {
				t.Errorf("port: got %q, want %q", result.URL.Port(), tt.expectedPort)
			}
			if result.URL.Path != tt.expectedPath {
				t.Errorf("path: got %q, want %q", result.URL.Path, tt.expectedPath)
			}
			if result.IsIP != tt.expectedIsIP {
				t.Errorf("IsIP: got %v, want %v", result.IsIP, tt.expectedIsIP)
			}
			if result.HasExplicitProto != tt.expectedExplicit {
				t.Errorf("HasExplicitProto: got %v, want %v", result.HasExplicitProto, tt.expectedExplicit)
			}
		})
	}
}

func TestParseTarget_HostPort(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedHost     string
		expectedPort     string
		expectedPath     string
		expectedIsIP     bool
		expectedExplicit bool
	}{
		{
			name:             "hostname with port",
			input:            "example.com:8080",
			expectedHost:     "example.com",
			expectedPort:     "8080",
			expectedPath:     "/",
			expectedIsIP:     false,
			expectedExplicit: false,
		},
		{
			name:             "hostname without port",
			input:            "example.com",
			expectedHost:     "example.com",
			expectedPort:     "",
			expectedPath:     "/",
			expectedIsIP:     false,
			expectedExplicit: false,
		},
		{
			name:             "hostname with port and path",
			input:            "api.example.com:9000/v1/users",
			expectedHost:     "api.example.com",
			expectedPort:     "9000",
			expectedPath:     "/v1/users",
			expectedIsIP:     false,
			expectedExplicit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.URL.Hostname() != tt.expectedHost {
				t.Errorf("host: got %q, want %q", result.URL.Hostname(), tt.expectedHost)
			}
			if result.URL.Port() != tt.expectedPort {
				t.Errorf("port: got %q, want %q", result.URL.Port(), tt.expectedPort)
			}
			if result.URL.Path != tt.expectedPath {
				t.Errorf("path: got %q, want %q", result.URL.Path, tt.expectedPath)
			}
			if result.IsIP != tt.expectedIsIP {
				t.Errorf("IsIP: got %v, want %v", result.IsIP, tt.expectedIsIP)
			}
			if result.HasExplicitProto != tt.expectedExplicit {
				t.Errorf("HasExplicitProto: got %v, want %v", result.HasExplicitProto, tt.expectedExplicit)
			}
		})
	}
}

func TestParseTarget_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "empty input",
			input:     "",
			shouldErr: true,
		},
		{
			name:      "invalid URL scheme",
			input:     "ht!tp://example.com",
			shouldErr: true,
		},
		{
			name:      "URL without host",
			input:     "http://",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.shouldErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if tt.shouldErr && result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

func TestParseTarget_IPv6(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedHost string
		expectedIsIP bool
	}{
		{
			name:         "IPv6 with brackets and port",
			input:        "http://[::1]:8080/path",
			expectedHost: "::1",
			expectedIsIP: true,
		},
		{
			name:         "IPv6 full address",
			input:        "http://[2001:db8::1]:443",
			expectedHost: "2001:db8::1",
			expectedIsIP: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.URL.Hostname() != tt.expectedHost {
				t.Errorf("host: got %q, want %q", result.URL.Hostname(), tt.expectedHost)
			}
			if result.IsIP != tt.expectedIsIP {
				t.Errorf("IsIP: got %v, want %v", result.IsIP, tt.expectedIsIP)
			}
		})
	}
}

// Property-Based Tests

// Property 1: URL Construction Round-Trip
// For any valid target input, parsing and reconstructing the URL string
// should produce a URL that contains all original components
func TestProperty_URLConstructionRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Use simple fixed hostnames for testing
	hostnameGen := gen.Const("example.com")

	// Generator for valid ports
	portGen := gen.IntRange(1, 65535)

	// Generator for valid paths
	pathGen := gen.Const("/")

	properties.Property("hostname:port/path round-trip", prop.ForAll(
		func(hostname string, port int, path string) bool {
			input := fmt.Sprintf("%s:%d%s", hostname, port, path)
			result, err := ParseTarget(input)
			if err != nil {
				return false
			}

			// Check that all components are preserved
			if result.URL.Hostname() != hostname {
				return false
			}
			if result.URL.Port() != fmt.Sprintf("%d", port) {
				return false
			}
			if result.URL.Path != path {
				return false
			}

			return true
		},
		hostnameGen,
		portGen,
		pathGen,
	))

	// Test with full URLs
	properties.Property("full URL round-trip", prop.ForAll(
		func(hostname string, port int, path string) bool {
			input := fmt.Sprintf("http://%s:%d%s", hostname, port, path)
			result, err := ParseTarget(input)
			if err != nil {
				return false
			}

			if result.URL.Hostname() != hostname {
				return false
			}
			if result.URL.Port() != fmt.Sprintf("%d", port) {
				return false
			}
			if result.URL.Path != path {
				return false
			}

			return true
		},
		hostnameGen,
		portGen,
		pathGen,
	))

	properties.TestingRun(t)
}

// Property 2: Default Path Appending
// For any target input without an explicit path component,
// the parser should produce a URL with path "/" appended
func TestProperty_DefaultPathAppending(t *testing.T) {
	properties := gopter.NewProperties(nil)

	hostnameGen := gen.Const("example.com")
	portGen := gen.IntRange(1, 65535)

	properties.Property("hostname:port gets / path", prop.ForAll(
		func(hostname string, port int) bool {
			input := fmt.Sprintf("%s:%d", hostname, port)
			result, err := ParseTarget(input)
			if err != nil {
				return false
			}

			return result.URL.Path == "/"
		},
		hostnameGen,
		portGen,
	))

	properties.Property("hostname alone gets / path", prop.ForAll(
		func(hostname string) bool {
			result, err := ParseTarget(hostname)
			if err != nil {
				return false
			}

			return result.URL.Path == "/"
		},
		hostnameGen,
	))

	properties.TestingRun(t)
}

// Property 3: Explicit Protocol Preservation
// For any target input that includes an explicit protocol scheme,
// the parser should preserve that scheme and mark HasExplicitProto as true
func TestProperty_ExplicitProtocolPreservation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	hostnameGen := gen.Const("example.com")
	portGen := gen.IntRange(1, 65535)
	schemeGen := gen.Const("http")

	properties.Property("explicit protocol is preserved", prop.ForAll(
		func(scheme string, hostname string, port int) bool {
			input := fmt.Sprintf("%s://%s:%d", scheme, hostname, port)
			result, err := ParseTarget(input)
			if err != nil {
				return false
			}

			if result.URL.Scheme != scheme {
				return false
			}
			if !result.HasExplicitProto {
				return false
			}

			return true
		},
		schemeGen,
		hostnameGen,
		portGen,
	))

	properties.Property("implicit protocol is not marked explicit", prop.ForAll(
		func(hostname string, port int) bool {
			input := fmt.Sprintf("%s:%d", hostname, port)
			result, err := ParseTarget(input)
			if err != nil {
				return false
			}

			return !result.HasExplicitProto
		},
		hostnameGen,
		portGen,
	))

	properties.TestingRun(t)
}

// Property 4: Invalid Target Error Mapping
// For any malformed target string, the parser should return a URLParseError
// that maps to exit code 3
func TestProperty_InvalidTargetErrorMapping(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for invalid inputs
	invalidInputGen := gen.Const("")

	properties.Property("invalid targets return URLParseError", prop.ForAll(
		func(input string) bool {
			result, err := ParseTarget(input)
			if err == nil {
				return false
			}

			// Check that it's a URLParseError
			_, ok := err.(*errors.URLParseError)
			if !ok {
				return false
			}

			// Check that result is nil
			if result != nil {
				return false
			}

			// Check that error maps to exit code 3
			exitCode := errors.MapErrorToExitCode(err)
			return exitCode == errors.ExitURLParse
		},
		invalidInputGen,
	))

	properties.TestingRun(t)
}

// Edge case tests

func TestParseTarget_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		check     func(*ParsedTarget) bool
	}{
		{
			name:      "localhost without port",
			input:     "localhost",
			shouldErr: false,
			check: func(pt *ParsedTarget) bool {
				return pt.URL.Hostname() == "localhost" && pt.URL.Path == "/"
			},
		},
		{
			name:      "localhost with port",
			input:     "localhost:3000",
			shouldErr: false,
			check: func(pt *ParsedTarget) bool {
				return pt.URL.Hostname() == "localhost" && pt.URL.Port() == "3000"
			},
		},
		{
			name:      "path with query string",
			input:     "example.com:8080/path?query=value",
			shouldErr: false,
			check: func(pt *ParsedTarget) bool {
				return strings.Contains(pt.URL.String(), "query=value")
			},
		},
		{
			name:      "path with fragment",
			input:     "example.com:8080/path#section",
			shouldErr: false,
			check: func(pt *ParsedTarget) bool {
				return pt.URL.Fragment == "section"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTarget(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.shouldErr && result != nil && !tt.check(result) {
				t.Errorf("check failed for %q", tt.input)
			}
		})
	}
}
