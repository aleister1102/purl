package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/user/purl/internal/cli"
	"github.com/user/purl/internal/errors"
	"github.com/user/purl/internal/target"
)

// NewTransport creates a configured http.Transport with TLS and timeout settings
// isIP indicates whether the target is an IP address (affects InsecureSkipVerify default)
func NewTransport(opts *cli.Options, parsedTarget *target.ParsedTarget) (*http.Transport, error) {
	// Create base transport
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: opts.ConnectTimeout,
		}).Dial,
		TLSHandshakeTimeout: opts.ConnectTimeout,
	}

	// Configure TLS settings
	tlsConfig := &tls.Config{}

	// Determine InsecureSkipVerify based on target type and flags
	// Default behavior: IP addresses skip verification unless --strict-ssl is set
	if parsedTarget.IsIP && !opts.StrictSSL {
		tlsConfig.InsecureSkipVerify = true
	} else if opts.Insecure {
		// -k/--insecure flag always skips verification
		tlsConfig.InsecureSkipVerify = true
	}

	// Get host for error messages
	host := "unknown"
	if parsedTarget != nil && parsedTarget.URL != nil {
		host = parsedTarget.URL.Host
	}

	// Load CA certificate if provided
	if opts.CACert != "" {
		caCert, err := os.ReadFile(opts.CACert)
		if err != nil {
			return nil, &errors.TLSError{
				Host:  host,
				Cause: fmt.Errorf("failed to read CA certificate: %w", err),
			}
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, &errors.TLSError{
				Host:  host,
				Cause: fmt.Errorf("failed to parse CA certificate"),
			}
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided
	if opts.Cert != "" && opts.Key != "" {
		cert, err := tls.LoadX509KeyPair(opts.Cert, opts.Key)
		if err != nil {
			return nil, &errors.TLSError{
				Host:  host,
				Cause: fmt.Errorf("failed to load client certificate: %w", err),
			}
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if opts.Cert != "" {
		return nil, &errors.TLSError{
			Host:  host,
			Cause: fmt.Errorf("--cert requires --key to be specified"),
		}
	} else if opts.Key != "" {
		return nil, &errors.TLSError{
			Host:  host,
			Cause: fmt.Errorf("--key requires --cert to be specified"),
		}
	}

	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

// ParseTimeout parses a duration string and returns a time.Duration
// Supports Go duration format (e.g., "10s", "5m", "100ms")
func ParseTimeout(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}

	return duration, nil
}

// ApplyTimeouts applies timeout settings from Options
// Returns the effective timeout duration (for context deadline)
// If no timeout is specified, returns the default 10 seconds
func ApplyTimeouts(opts *cli.Options) time.Duration {
	// --max-time is an alias for --timeout
	if opts.Timeout > 0 {
		return opts.Timeout
	}

	// Default to 10 seconds if no timeout specified
	return 10 * time.Second
}

// GetConnectTimeout returns the connect timeout duration
// If not specified, returns the default 10 seconds
func GetConnectTimeout(opts *cli.Options) time.Duration {
	if opts.ConnectTimeout > 0 {
		return opts.ConnectTimeout
	}

	// Default to 10 seconds if not specified
	return 10 * time.Second
}
