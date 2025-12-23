package transport

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/user/purl/internal/cli"
	"github.com/user/purl/internal/target"
)

// Unit Tests

func TestNewTransport_IPAddressInsecureSkipVerify(t *testing.T) {
	tests := []struct {
		name               string
		isIP               bool
		strictSSL          bool
		insecure           bool
		expectedSkipVerify bool
	}{
		{
			name:               "IP address without flags defaults to skip verify",
			isIP:               true,
			strictSSL:          false,
			insecure:           false,
			expectedSkipVerify: true,
		},
		{
			name:               "IP address with --strict-ssl enforces verification",
			isIP:               true,
			strictSSL:          true,
			insecure:           false,
			expectedSkipVerify: false,
		},
		{
			name:               "IP address with -k/--insecure skips verification",
			isIP:               true,
			strictSSL:          false,
			insecure:           true,
			expectedSkipVerify: true,
		},
		{
			name:               "hostname without flags enforces verification",
			isIP:               false,
			strictSSL:          false,
			insecure:           false,
			expectedSkipVerify: false,
		},
		{
			name:               "hostname with -k/--insecure skips verification",
			isIP:               false,
			strictSSL:          false,
			insecure:           true,
			expectedSkipVerify: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &cli.Options{
				StrictSSL: tt.strictSSL,
				Insecure:  tt.insecure,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: tt.isIP,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if transport.TLSClientConfig.InsecureSkipVerify != tt.expectedSkipVerify {
				t.Errorf("InsecureSkipVerify: got %v, want %v",
					transport.TLSClientConfig.InsecureSkipVerify,
					tt.expectedSkipVerify)
			}
		})
	}
}

func TestNewTransport_TimeoutSettings(t *testing.T) {
	tests := []struct {
		name                string
		connectTimeout      time.Duration
		expectedDialTimeout time.Duration
		expectedTLSTimeout  time.Duration
	}{
		{
			name:                "default connect timeout",
			connectTimeout:      0,
			expectedDialTimeout: 0, // Will be set by GetConnectTimeout
			expectedTLSTimeout:  0,
		},
		{
			name:                "custom connect timeout",
			connectTimeout:      5 * time.Second,
			expectedDialTimeout: 5 * time.Second,
			expectedTLSTimeout:  5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &cli.Options{
				ConnectTimeout: tt.connectTimeout,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: false,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if transport.TLSHandshakeTimeout != tt.expectedTLSTimeout {
				t.Errorf("TLSHandshakeTimeout: got %v, want %v",
					transport.TLSHandshakeTimeout,
					tt.expectedTLSTimeout)
			}
		})
	}
}

func TestParseTimeout_ValidDurations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "seconds",
			input:    "10s",
			expected: 10 * time.Second,
		},
		{
			name:     "milliseconds",
			input:    "500ms",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "minutes",
			input:    "2m",
			expected: 2 * time.Minute,
		},
		{
			name:     "microseconds",
			input:    "1000µs",
			expected: 1000 * time.Microsecond,
		},
		{
			name:     "nanoseconds",
			input:    "1000000ns",
			expected: 1000000 * time.Nanosecond,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeout(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTimeout_InvalidDurations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid format",
			input: "not-a-duration",
		},
		{
			name:  "missing unit",
			input: "10",
		},
		{
			name:  "invalid unit",
			input: "10x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeout(tt.input)
			if err == nil {
				t.Errorf("expected error, got nil (result: %v)", result)
			}
		})
	}
}

func TestApplyTimeouts_DefaultTimeout(t *testing.T) {
	opts := &cli.Options{
		Timeout: 0,
	}

	result := ApplyTimeouts(opts)
	expected := 10 * time.Second

	if result != expected {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestApplyTimeouts_CustomTimeout(t *testing.T) {
	opts := &cli.Options{
		Timeout: 30 * time.Second,
	}

	result := ApplyTimeouts(opts)
	expected := 30 * time.Second

	if result != expected {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestGetConnectTimeout_DefaultTimeout(t *testing.T) {
	opts := &cli.Options{
		ConnectTimeout: 0,
	}

	result := GetConnectTimeout(opts)
	expected := 10 * time.Second

	if result != expected {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestGetConnectTimeout_CustomTimeout(t *testing.T) {
	opts := &cli.Options{
		ConnectTimeout: 5 * time.Second,
	}

	result := GetConnectTimeout(opts)
	expected := 5 * time.Second

	if result != expected {
		t.Errorf("got %v, want %v", result, expected)
	}
}

// Property-Based Tests

// Property 5: Protocol Override Behavior
// For any --proto flag value ("http" or "https"), the protocol detector
// should use exactly that protocol without attempting the other
// Note: This property is tested in protocol detector, but we verify
// that transport doesn't interfere with protocol selection
func TestProperty_TransportCreationSucceeds(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for boolean flags
	boolGen := gen.Bool()

	properties.Property("transport creation succeeds for all flag combinations", prop.ForAll(
		func(isIP bool, strictSSL bool, insecure bool) bool {
			opts := &cli.Options{
				StrictSSL: strictSSL,
				Insecure:  insecure,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: isIP,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				return false
			}

			// Verify transport is properly configured
			if transport == nil {
				return false
			}
			if transport.TLSClientConfig == nil {
				return false
			}

			return true
		},
		boolGen,
		boolGen,
		boolGen,
	))

	properties.TestingRun(t)
}

// Property 6: TLS InsecureSkipVerify Configuration
// For any target that is an IP address:
// - Without --strict-ssl: InsecureSkipVerify SHALL be true
// - With --strict-ssl: InsecureSkipVerify SHALL be false
// - With -k/--insecure: InsecureSkipVerify SHALL be true regardless of target type
func TestProperty_TLSInsecureSkipVerifyConfiguration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("IP address without flags skips verification", prop.ForAll(
		func() bool {
			opts := &cli.Options{
				StrictSSL: false,
				Insecure:  false,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: true,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				return false
			}

			return transport.TLSClientConfig.InsecureSkipVerify == true
		},
	))

	properties.Property("IP address with --strict-ssl enforces verification", prop.ForAll(
		func() bool {
			opts := &cli.Options{
				StrictSSL: true,
				Insecure:  false,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: true,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				return false
			}

			return transport.TLSClientConfig.InsecureSkipVerify == false
		},
	))

	properties.Property("any target with -k/--insecure skips verification", prop.ForAll(
		func(isIP bool) bool {
			opts := &cli.Options{
				StrictSSL: false,
				Insecure:  true,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: isIP,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				return false
			}

			return transport.TLSClientConfig.InsecureSkipVerify == true
		},
		gen.Bool(),
	))

	properties.Property("hostname without flags enforces verification", prop.ForAll(
		func() bool {
			opts := &cli.Options{
				StrictSSL: false,
				Insecure:  false,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: false,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if err != nil {
				return false
			}

			return transport.TLSClientConfig.InsecureSkipVerify == false
		},
	))

	properties.TestingRun(t)
}

// Property 18: Timeout Duration Parsing
// For any --timeout, --connect-timeout, or --max-time flag with a valid Go duration string,
// the transport should parse it to the correct time.Duration value
func TestProperty_TimeoutDurationParsing(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for valid duration values in seconds
	secondsGen := gen.IntRange(1, 300)

	properties.Property("timeout parsing produces correct duration", prop.ForAll(
		func(seconds int) bool {
			input := fmt.Sprintf("%ds", seconds)
			result, err := ParseTimeout(input)
			if err != nil {
				return false
			}

			expected := time.Duration(seconds) * time.Second
			return result == expected
		},
		secondsGen,
	))

	properties.Property("millisecond timeout parsing", prop.ForAll(
		func(millis int) bool {
			if millis < 1 || millis > 300000 {
				return true // Skip invalid ranges
			}

			input := fmt.Sprintf("%dms", millis)
			result, err := ParseTimeout(input)
			if err != nil {
				return false
			}

			expected := time.Duration(millis) * time.Millisecond
			return result == expected
		},
		gen.IntRange(1, 300000),
	))

	properties.Property("empty string returns zero duration", prop.ForAll(
		func() bool {
			result, err := ParseTimeout("")
			if err != nil {
				return false
			}

			return result == 0
		},
	))

	properties.TestingRun(t)
}

// Edge case tests

func TestNewTransport_CertificateValidation(t *testing.T) {
	tests := []struct {
		name      string
		cert      string
		key       string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "cert without key",
			cert:      "/path/to/cert",
			key:       "",
			shouldErr: true,
			errMsg:    "requires --key",
		},
		{
			name:      "key without cert",
			cert:      "",
			key:       "/path/to/key",
			shouldErr: true,
			errMsg:    "requires --cert",
		},
		{
			name:      "no cert or key",
			cert:      "",
			key:       "",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &cli.Options{
				Cert: tt.cert,
				Key:  tt.key,
			}

			parsedTarget := &target.ParsedTarget{
				IsIP: false,
			}

			transport, err := NewTransport(opts, parsedTarget)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if transport == nil {
					t.Errorf("expected transport, got nil")
				}
			}
		})
	}
}

func TestNewTransport_DialerConfiguration(t *testing.T) {
	opts := &cli.Options{
		ConnectTimeout: 5 * time.Second,
	}

	parsedTarget := &target.ParsedTarget{
		IsIP: false,
	}

	transport, err := NewTransport(opts, parsedTarget)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that the transport has a Dial function configured
	if transport.Dial == nil {
		t.Errorf("expected Dial function to be configured")
	}

	// Verify TLS handshake timeout is set
	if transport.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLSHandshakeTimeout: got %v, want %v",
			transport.TLSHandshakeTimeout,
			5*time.Second)
	}
}

func TestNewTransport_TLSConfigExists(t *testing.T) {
	opts := &cli.Options{}

	parsedTarget := &target.ParsedTarget{
		IsIP: false,
	}

	transport, err := NewTransport(opts, parsedTarget)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transport.TLSClientConfig == nil {
		t.Errorf("expected TLSClientConfig to be configured")
	}
}

func TestParseTimeout_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Duration
		shouldErr bool
	}{
		{
			name:      "zero duration",
			input:     "0s",
			expected:  0,
			shouldErr: false,
		},
		{
			name:      "very large duration",
			input:     "999999s",
			expected:  999999 * time.Second,
			shouldErr: false,
		},
		{
			name:      "fractional seconds",
			input:     "1.5s",
			expected:  1500 * time.Millisecond,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeout(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// Test that transport can be used with net.Dialer
func TestNewTransport_DialerUsable(t *testing.T) {
	opts := &cli.Options{
		ConnectTimeout: 5 * time.Second,
	}

	parsedTarget := &target.ParsedTarget{
		IsIP: false,
	}

	transport, err := NewTransport(opts, parsedTarget)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we can create a dialer from the transport settings
	if transport.Dial == nil {
		t.Errorf("expected Dial function to be set")
	}

	// Try to use the dialer (this will fail to connect but should not panic)
	// We're just testing that the configuration is valid
	conn, err := transport.Dial("tcp", "localhost:99999")
	if err == nil {
		conn.Close()
	}
	// We expect an error here since we're connecting to an invalid port
	// The important thing is that it doesn't panic
}
