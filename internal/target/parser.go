package target

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/aleister1102/purl/internal/errors"
)

// ParsedTarget represents a normalized URL with metadata
type ParsedTarget struct {
	URL              *url.URL
	IsIP             bool
	HasExplicitProto bool
	OriginalInput    string
}

// ParseTarget normalizes various input formats to a URL
// Supports formats:
// - IP:PORT/path (e.g., 192.168.1.1:8080/api)
// - host:PORT (e.g., example.com:443)
// - full URL (e.g., http://example.com:8080/path)
func ParseTarget(input string) (*ParsedTarget, error) {
	if input == "" {
		return nil, &errors.URLParseError{
			Input:   input,
			Message: "target cannot be empty",
		}
	}

	result := &ParsedTarget{
		OriginalInput: input,
	}

	// Check if input already has a scheme (http://, https://, etc.)
	if strings.Contains(input, "://") {
		result.HasExplicitProto = true
		parsedURL, err := url.Parse(input)
		if err != nil {
			return nil, &errors.URLParseError{
				Input:   input,
				Message: fmt.Sprintf("invalid URL: %v", err),
			}
		}

		// Validate that the URL has a host
		if parsedURL.Host == "" {
			return nil, &errors.URLParseError{
				Input:   input,
				Message: "URL must contain a host",
			}
		}

		result.URL = parsedURL
		result.IsIP = isIPAddress(parsedURL.Hostname())
		return result, nil
	}

	// No scheme provided - need to parse as IP:PORT/path or host:PORT
	// First, separate the path from the host:port part
	var hostPort string
	var path string

	// Find the first slash to separate host:port from path
	slashIdx := strings.Index(input, "/")
	if slashIdx != -1 {
		hostPort = input[:slashIdx]
		path = input[slashIdx:]
	} else {
		hostPort = input
		path = "/"
	}

	// Parse the host:port part
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		// No port specified, treat entire hostPort as host
		host = hostPort
		port = ""
	}

	// Validate host is not empty
	if host == "" {
		return nil, &errors.URLParseError{
			Input:   input,
			Message: "target must contain a host",
		}
	}

	// Determine if this is an IP address
	result.IsIP = isIPAddress(host)

	// Construct the URL with http scheme (protocol detection happens later)
	var urlStr string
	if port != "" {
		urlStr = fmt.Sprintf("http://%s:%s%s", host, port, path)
	} else {
		urlStr = fmt.Sprintf("http://%s%s", host, path)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, &errors.URLParseError{
			Input:   input,
			Message: fmt.Sprintf("invalid target format: %v", err),
		}
	}

	result.URL = parsedURL
	result.HasExplicitProto = false
	return result, nil
}

// isIPAddress checks if a string is a valid IP address (v4 or v6)
func isIPAddress(host string) bool {
	// Remove brackets for IPv6 addresses
	testHost := strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	return net.ParseIP(testHost) != nil
}
