package request

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/user/purl/internal/cli"
	"github.com/user/purl/internal/target"
)

// BuildRequest creates an http.Request from CLI options and parsed target
// Handles method override, headers, body, authentication, and special flags
func BuildRequest(ctx context.Context, parsedTarget *target.ParsedTarget, opts *cli.Options) (*http.Request, error) {
	// Determine the HTTP method
	method := opts.Method
	if method == "" {
		// Default to GET, unless data is provided
		if opts.Data != "" || opts.DataRaw != "" {
			method = "POST"
		} else if opts.Head {
			method = "HEAD"
		} else {
			method = "GET"
		}
	}

	// Prepare the request body
	var body io.Reader
	if opts.DataRaw != "" {
		// --data-raw: preserve special characters literally
		body = strings.NewReader(opts.DataRaw)
	} else if opts.Data != "" {
		// -d/--data: use as-is (no special interpretation in this implementation)
		body = strings.NewReader(opts.Data)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, parsedTarget.URL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers from -H flags
	for _, header := range opts.Headers {
		// Parse header as "Name: Value"
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(name, value)
		}
	}

	// Add authentication header if -u flag is provided
	if opts.User != "" {
		addBasicAuth(req, opts.User)
	}

	// Add Cookie header if --cookie flag is provided
	if opts.Cookie != "" {
		req.Header.Set("Cookie", opts.Cookie)
	}

	// Add User-Agent header if --user-agent flag is provided
	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	}

	// Add Referer header if --referer flag is provided
	if opts.Referer != "" {
		req.Header.Set("Referer", opts.Referer)
	}

	// Add JSON headers if --json flag is provided
	if opts.JSON {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	return req, nil
}

// addBasicAuth adds Basic Authentication header to the request
// Expects user string in format "user:password"
func addBasicAuth(req *http.Request, user string) {
	// Encode user:password in base64
	encoded := base64.StdEncoding.EncodeToString([]byte(user))
	req.Header.Set("Authorization", "Basic "+encoded)
}
