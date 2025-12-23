package request

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/user/curlz/internal/cli"
	"github.com/user/curlz/internal/target"
)

// Unit Tests

func TestBuildRequest_BasicGET(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
		IsIP:             false,
		HasExplicitProto: false,
	}

	opts := &cli.Options{}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method GET, got %s", req.Method)
	}

	if req.URL.String() != "http://example.com/" {
		t.Errorf("Expected URL http://example.com/, got %s", req.URL.String())
	}
}

func TestBuildRequest_MethodOverride(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Method: "DELETE",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Method != "DELETE" {
		t.Errorf("Expected method DELETE, got %s", req.Method)
	}
}

func TestBuildRequest_DataDefaultsToPost(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Data: "key=value",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("Expected method POST when data is provided, got %s", req.Method)
	}
}

func TestBuildRequest_DataRawPreservesSpecialChars(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		DataRaw: "key=value&special=@#$%",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	body := make([]byte, len(opts.DataRaw))
	n, err := req.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body[:n]) != opts.DataRaw {
		t.Errorf("Expected body %s, got %s", opts.DataRaw, string(body[:n]))
	}
}

func TestBuildRequest_HeadersAccumulate(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Headers: []string{
			"X-Custom: value1",
			"X-Another: value2",
			"Content-Type: application/json",
		},
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("X-Custom") != "value1" {
		t.Errorf("Expected X-Custom header value1, got %s", req.Header.Get("X-Custom"))
	}

	if req.Header.Get("X-Another") != "value2" {
		t.Errorf("Expected X-Another header value2, got %s", req.Header.Get("X-Another"))
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header application/json, got %s", req.Header.Get("Content-Type"))
	}
}

func TestBuildRequest_BasicAuth(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		User: "user:password",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:password"))

	if authHeader != expected {
		t.Errorf("Expected Authorization header %s, got %s", expected, authHeader)
	}
}

func TestBuildRequest_CookieHeader(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Cookie: "session=abc123; path=/",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("Cookie") != "session=abc123; path=/" {
		t.Errorf("Expected Cookie header, got %s", req.Header.Get("Cookie"))
	}
}

func TestBuildRequest_UserAgentHeader(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		UserAgent: "CustomAgent/1.0",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("User-Agent") != "CustomAgent/1.0" {
		t.Errorf("Expected User-Agent header CustomAgent/1.0, got %s", req.Header.Get("User-Agent"))
	}
}

func TestBuildRequest_RefererHeader(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Referer: "http://referrer.com",
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("Referer") != "http://referrer.com" {
		t.Errorf("Expected Referer header http://referrer.com, got %s", req.Header.Get("Referer"))
	}
}

func TestBuildRequest_HeadMethod(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		Head: true,
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Method != "HEAD" {
		t.Errorf("Expected method HEAD, got %s", req.Method)
	}
}

func TestBuildRequest_JSONHeaders(t *testing.T) {
	parsedTarget := &target.ParsedTarget{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/",
		},
	}

	opts := &cli.Options{
		JSON: true,
	}

	req, err := BuildRequest(context.Background(), parsedTarget, opts)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Expected Accept application/json, got %s", req.Header.Get("Accept"))
	}
}

// Property-Based Tests

// Property 8: HTTP Method Setting
// For any -X/--request flag value, the Request_Builder SHALL set the request method to exactly that value
func TestProperty_HTTPMethodSetting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("HTTP Method Setting - Feature: curlz-http-probe, Property 8: HTTP Method Setting",
		prop.ForAll(
			func(method string) bool {
				// Skip empty methods
				if method == "" {
					return true
				}

				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					Method: method,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Method == method
			},
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// Property 9: Header Accumulation
// For any sequence of -H/--header flags, the Request_Builder SHALL include ALL specified headers
func TestProperty_HeaderAccumulation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Header Accumulation - Feature: curlz-http-probe, Property 9: Header Accumulation",
		prop.ForAll(
			func(headers []string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				// Create valid headers with unique names to avoid overwrites
				validHeaders := []string{}
				for i, h := range headers {
					if h != "" {
						validHeaders = append(validHeaders, fmt.Sprintf("X-Test-%d: %s", i, h))
					}
				}

				opts := &cli.Options{
					Headers: validHeaders,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				// Check that all headers are present
				for _, h := range validHeaders {
					parts := strings.SplitN(h, ":", 2)
					if len(parts) == 2 {
						name := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						if req.Header.Get(name) != value {
							return false
						}
					}
				}

				return true
			},
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.TestingRun(t)
}

// Property 10: Data Flag Behavior
// For any -d/--data flag value, the Request_Builder SHALL set the request body and default method to POST
func TestProperty_DataFlagBehavior(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Data Flag Behavior - Feature: curlz-http-probe, Property 10: Data Flag Behavior",
		prop.ForAll(
			func(data string) bool {
				// Only test non-empty data (empty data is indistinguishable from no flag)
				if data == "" {
					return true
				}

				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					Data: data,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				// Method should default to POST when data is provided
				return req.Method == "POST"
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}

// Property 11: Raw Data Preservation
// For any --data-raw flag value containing special characters, the Request_Builder SHALL preserve those characters literally
func TestProperty_RawDataPreservation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Raw Data Preservation - Feature: curlz-http-probe, Property 11: Raw Data Preservation",
		prop.ForAll(
			func(data string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					DataRaw: data,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				// Read the body and verify it matches the raw data
				if req.Body == nil {
					return data == ""
				}

				body := make([]byte, len(data)+1)
				n, err := req.Body.Read(body)
				if err != nil && err.Error() != "EOF" {
					return false
				}

				return string(body[:n]) == data
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}

// Property 12: Basic Auth Header Generation
// For any -u/--user flag value in format "user:password", the Request_Builder SHALL set the Authorization header correctly
func TestProperty_BasicAuthHeaderGeneration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Basic Auth Header Generation - Feature: curlz-http-probe, Property 12: Basic Auth Header Generation",
		prop.ForAll(
			func(user, password string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				userStr := user + ":" + password
				opts := &cli.Options{
					User: userStr,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				authHeader := req.Header.Get("Authorization")
				expected := "Basic " + base64.StdEncoding.EncodeToString([]byte(userStr))

				return authHeader == expected
			},
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// Property 13: Cookie Header Setting
// For any --cookie flag value, the Request_Builder SHALL set the Cookie header to exactly that value
func TestProperty_CookieHeaderSetting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Cookie Header Setting - Feature: curlz-http-probe, Property 13: Cookie Header Setting",
		prop.ForAll(
			func(cookie string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					Cookie: cookie,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Header.Get("Cookie") == cookie
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}

// Property 14: User-Agent Header Setting
// For any --user-agent flag value, the Request_Builder SHALL set the User-Agent header to exactly that value
func TestProperty_UserAgentHeaderSetting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("User-Agent Header Setting - Feature: curlz-http-probe, Property 14: User-Agent Header Setting",
		prop.ForAll(
			func(ua string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					UserAgent: ua,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Header.Get("User-Agent") == ua
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}

// Property 15: Referer Header Setting
// For any --referer flag value, the Request_Builder SHALL set the Referer header to exactly that value
func TestProperty_RefererHeaderSetting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Referer Header Setting - Feature: curlz-http-probe, Property 15: Referer Header Setting",
		prop.ForAll(
			func(referer string) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					Referer: referer,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Header.Get("Referer") == referer
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}

// Property 16: HEAD Request Flag
// For any request with -I/--head flag, the Request_Builder SHALL set the method to "HEAD"
func TestProperty_HEADRequestFlag(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("HEAD Request Flag - Feature: curlz-http-probe, Property 16: HEAD Request Flag",
		prop.ForAll(
			func(_ struct{}) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					Head: true,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Method == "HEAD"
			},
			gen.Const(struct{}{}),
		),
	)

	properties.TestingRun(t)
}

// Property 17: JSON Flag Headers
// For any request with --json flag, the Request_Builder SHALL set both Content-Type and Accept headers to application/json
func TestProperty_JSONFlagHeaders(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("JSON Flag Headers - Feature: curlz-http-probe, Property 17: JSON Flag Headers",
		prop.ForAll(
			func(_ struct{}) bool {
				parsedTarget := &target.ParsedTarget{
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/",
					},
				}

				opts := &cli.Options{
					JSON: true,
				}

				req, err := BuildRequest(context.Background(), parsedTarget, opts)
				if err != nil {
					return false
				}

				return req.Header.Get("Content-Type") == "application/json" &&
					req.Header.Get("Accept") == "application/json"
			},
			gen.Const(struct{}{}),
		),
	)

	properties.TestingRun(t)
}
