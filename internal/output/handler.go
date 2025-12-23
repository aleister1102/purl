package output

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/user/curlz/internal/cli"
	"github.com/user/curlz/internal/protocol"
)

// Handler manages output formatting and writing
type Handler struct {
	opts *cli.Options
}

// NewHandler creates a new output handler
func NewHandler(opts *cli.Options) *Handler {
	return &Handler{
		opts: opts,
	}
}

// WriteResponse handles writing the response to stdout or file, with optional verbose output
func (h *Handler) WriteResponse(req *http.Request, result *protocol.ProbeResult) error {
	// Print verbose request details to stderr if requested
	if h.opts.Verbose {
		if err := h.printVerboseRequest(req); err != nil {
			return err
		}
	}

	// Print status line to stdout
	if err := h.printStatusLine(result); err != nil {
		return err
	}

	// Print verbose response headers to stderr if requested
	if h.opts.Verbose && result.Response != nil {
		if err := h.printVerboseResponse(result.Response); err != nil {
			return err
		}
	}

	// Print TLS details to stderr if requested
	if h.opts.VerboseTLS && result.Response != nil {
		if err := h.printTLSDetails(result.Response); err != nil {
			return err
		}
	}

	// Stream response body to stdout or file
	if result.Response != nil && result.Response.Body != nil {
		if err := h.writeResponseBody(result.Response); err != nil {
			return err
		}
	}

	return nil
}

// printStatusLine prints the formatted status line
// Format: "[PROTO] Status: CODE Time: Xs"
func (h *Handler) printStatusLine(result *protocol.ProbeResult) error {
	proto := result.Protocol
	if proto == "" {
		proto = "HTTP"
	} else {
		proto = formatProto(proto)
	}

	statusCode := result.StatusCode
	if statusCode == 0 {
		statusCode = 0 // Will show as "0" if no response
	}

	// Format duration with appropriate unit
	durationStr := formatDuration(result.Duration)

	statusLine := fmt.Sprintf("[%s] Status: %d Time: %s\n", proto, statusCode, durationStr)
	_, err := fmt.Fprint(os.Stdout, statusLine)
	return err
}

// formatProto converts protocol string to uppercase
func formatProto(proto string) string {
	if proto == "http" {
		return "HTTP"
	}
	if proto == "https" {
		return "HTTPS"
	}
	return proto
}

// formatDuration formats a duration with appropriate unit
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	// Convert to seconds with appropriate precision
	seconds := d.Seconds()
	if seconds < 0.001 {
		// Microseconds
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	}
	if seconds < 1 {
		// Milliseconds
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	// Seconds
	if seconds == float64(int64(seconds)) {
		return fmt.Sprintf("%.0fs", seconds)
	}
	return fmt.Sprintf("%.2fs", seconds)
}

// writeResponseBody writes the response body to stdout or file
func (h *Handler) writeResponseBody(resp *http.Response) error {
	var writer io.Writer

	if h.opts.Output != "" {
		// Write to file
		file, err := os.Create(h.opts.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	} else {
		// Write to stdout
		writer = os.Stdout
	}

	// Copy response body to writer
	_, err := io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write response body: %w", err)
	}

	return nil
}

// printVerboseRequest prints request details to stderr
func (h *Handler) printVerboseRequest(req *http.Request) error {
	// Print request line
	fmt.Fprintf(os.Stderr, "> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)

	// Print request headers
	for name, values := range req.Header {
		for _, value := range values {
			// Mask Authorization header for security
			if name == "Authorization" {
				fmt.Fprintf(os.Stderr, "> %s: [REDACTED]\n", name)
			} else {
				fmt.Fprintf(os.Stderr, "> %s: %s\n", name, value)
			}
		}
	}

	// Print blank line after headers
	fmt.Fprintf(os.Stderr, ">\n")

	return nil
}

// printVerboseResponse prints response headers to stderr
func (h *Handler) printVerboseResponse(resp *http.Response) error {
	// Print status line
	fmt.Fprintf(os.Stderr, "< %s %d %s\n", resp.Proto, resp.StatusCode, http.StatusText(resp.StatusCode))

	// Print response headers
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "< %s: %s\n", name, value)
		}
	}

	// Print blank line after headers
	fmt.Fprintf(os.Stderr, "<\n")

	return nil
}

// printTLSDetails prints TLS handshake details to stderr
func (h *Handler) printTLSDetails(resp *http.Response) error {
	if resp.TLS == nil {
		fmt.Fprintf(os.Stderr, "* No TLS connection\n")
		return nil
	}

	tls := resp.TLS

	// Print TLS version
	tlsVersion := getTLSVersionString(tls.Version)
	fmt.Fprintf(os.Stderr, "* TLS Version: %s\n", tlsVersion)

	// Print cipher suite
	cipherSuite := getCipherSuiteName(tls.CipherSuite)
	fmt.Fprintf(os.Stderr, "* Cipher Suite: %s\n", cipherSuite)

	// Print certificate info
	if len(tls.PeerCertificates) > 0 {
		cert := tls.PeerCertificates[0]
		fmt.Fprintf(os.Stderr, "* Subject: %s\n", cert.Subject.String())
		fmt.Fprintf(os.Stderr, "* Issuer: %s\n", cert.Issuer.String())
		fmt.Fprintf(os.Stderr, "* Valid From: %s\n", cert.NotBefore.String())
		fmt.Fprintf(os.Stderr, "* Valid Until: %s\n", cert.NotAfter.String())
	}

	return nil
}

// getTLSVersionString converts TLS version constant to string
func getTLSVersionString(version uint16) string {
	switch version {
	case 0x0301:
		return "TLS 1.0"
	case 0x0302:
		return "TLS 1.1"
	case 0x0303:
		return "TLS 1.2"
	case 0x0304:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

// getCipherSuiteName converts cipher suite constant to name
func getCipherSuiteName(suite uint16) string {
	// Map common cipher suites
	cipherSuites := map[uint16]string{
		0x002f: "TLS_RSA_WITH_AES_128_CBC_SHA",
		0x0035: "TLS_RSA_WITH_AES_256_CBC_SHA",
		0x003c: "TLS_RSA_WITH_AES_128_CBC_SHA256",
		0x003d: "TLS_RSA_WITH_AES_256_CBC_SHA256",
		0x009c: "TLS_RSA_WITH_AES_128_GCM_SHA256",
		0x009d: "TLS_RSA_WITH_AES_256_GCM_SHA384",
		0x1301: "TLS_AES_128_GCM_SHA256",
		0x1302: "TLS_AES_256_GCM_SHA384",
		0x1303: "TLS_CHACHA20_POLY1305_SHA256",
		0xc02b: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		0xc02c: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		0xc02f: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		0xc030: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	}

	if name, ok := cipherSuites[suite]; ok {
		return name
	}

	return fmt.Sprintf("Unknown (0x%04x)", suite)
}
