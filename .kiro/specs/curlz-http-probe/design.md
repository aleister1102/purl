# Design Document: Curlz HTTP Probe

## Overview

Curlz is a Go CLI tool that serves as a drop-in curl replacement with automatic HTTP/HTTPS protocol detection. The architecture follows a modular design with clear separation between CLI parsing, URL construction, protocol detection, request building, and HTTP transport.

The tool prioritizes curl compatibility while adding intelligent protocol probing - attempting HTTP first with a quick timeout, then falling back to HTTPS. This makes it ideal for security testing and API exploration where the protocol may be unknown.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                           main.go                                │
│                    (Entry point, CLI setup)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Parser                                │
│              (urfave/cli/v2 + custom flag handling)              │
│  - Parses curl-compatible flags                                  │
│  - Extracts target URL from positional args                      │
│  - Returns Options struct                                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Target Parser                               │
│  - Normalizes IP:PORT/path → URL                                 │
│  - Detects explicit protocol vs auto-detect                      │
│  - Validates URL structure                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Protocol Detector                             │
│  - Auto mode: HTTP (3s) → HTTPS fallback (7s)                   │
│  - Manual mode: --proto http|https                               │
│  - Returns working protocol + response                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Request Builder                              │
│  - Constructs http.Request from Options                          │
│  - Sets headers, body, auth, cookies                             │
│  - Applies method override                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Transport                                  │
│  - Configures http.Transport with TLS settings                   │
│  - Handles timeouts via context                                  │
│  - Manages InsecureSkipVerify for IP targets                     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Output Handler                               │
│  - Streams response to stdout or file                            │
│  - Prints verbose info to stderr                                 │
│  - Formats status line                                           │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Options Struct

```go
// Options holds all parsed CLI flags and target information
type Options struct {
    // Target
    Target      string
    Proto       string // "auto", "http", "https"
    
    // Request
    Method      string
    Headers     []string
    Data        string
    DataRaw     string
    User        string // user:password
    Cookie      string
    UserAgent   string
    Referer     string
    
    // Output
    Verbose     bool
    VerboseTLS  bool
    Output      string
    Head        bool
    JSON        bool
    
    // TLS
    Insecure    bool
    CACert      string
    Cert        string
    Key         string
    StrictSSL   bool
    
    // Timeouts
    Timeout        time.Duration
    ConnectTimeout time.Duration
}
```

### Target Parser Interface

```go
// ParsedTarget represents a normalized URL with metadata
type ParsedTarget struct {
    URL           *url.URL
    IsIP          bool
    HasExplicitProto bool
    OriginalInput string
}

// ParseTarget normalizes various input formats to a URL
func ParseTarget(input string) (*ParsedTarget, error)
```

### Protocol Detector Interface

```go
// ProbeResult contains the result of protocol detection
type ProbeResult struct {
    Protocol    string        // "http" or "https"
    StatusCode  int
    Duration    time.Duration
    Response    *http.Response
    Error       error
}

// DetectProtocol probes the target and returns the working protocol
func DetectProtocol(target *ParsedTarget, opts *Options) (*ProbeResult, error)
```

### Request Builder Interface

```go
// BuildRequest creates an http.Request from options
func BuildRequest(ctx context.Context, target *ParsedTarget, opts *Options) (*http.Request, error)
```

### Transport Interface

```go
// NewTransport creates a configured http.Transport
func NewTransport(opts *Options, isIP bool) (*http.Transport, error)
```

### Exit Code Mapper

```go
// ExitCode maps errors to curl-compatible exit codes
func ExitCode(err error) int
```

## Data Models

### Error Types

```go
type URLParseError struct {
    Input   string
    Message string
}

type ConnectionError struct {
    Host    string
    Port    string
    Cause   error
}

type TimeoutError struct {
    Duration time.Duration
    Phase    string // "connect" or "request"
}

type TLSError struct {
    Host  string
    Cause error
}
```

### Exit Code Constants

```go
const (
    ExitSuccess       = 0
    ExitUnknownFlag   = 2
    ExitURLParse      = 3
    ExitNoRoute       = 6
    ExitConnectFailed = 7
    ExitTimeout       = 28
    ExitTLSError      = 35
)
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: URL Construction Round-Trip

*For any* valid target input (IP:PORT/path, host:PORT, or full URL), parsing the target and reconstructing the URL string SHALL produce a URL that contains all original components (host, port, path) in valid URL format.

**Validates: Requirements 1.1, 1.2, 1.3**

### Property 2: Default Path Appending

*For any* target input without an explicit path component, THE Target_Parser SHALL produce a URL with path "/" appended.

**Validates: Requirements 1.2**

### Property 3: Explicit Protocol Preservation

*For any* target input that includes an explicit protocol scheme (http:// or https://), THE Target_Parser SHALL preserve that scheme without modification and mark HasExplicitProto as true.

**Validates: Requirements 1.3**

### Property 4: Invalid Target Error Mapping

*For any* malformed target string that cannot be parsed as a valid URL, THE Target_Parser SHALL return a URLParseError that maps to exit code 3.

**Validates: Requirements 1.4, 7.2**

### Property 5: Protocol Override Behavior

*For any* --proto flag value ("http" or "https"), THE Protocol_Detector SHALL use exactly that protocol without attempting the other, regardless of target format.

**Validates: Requirements 2.5, 2.6**

### Property 6: TLS InsecureSkipVerify Configuration

*For any* target that is an IP address:
- Without --strict-ssl: InsecureSkipVerify SHALL be true
- With --strict-ssl: InsecureSkipVerify SHALL be false
- With -k/--insecure: InsecureSkipVerify SHALL be true regardless of target type

**Validates: Requirements 2.4, 5.1, 5.5, 5.6**

### Property 7: Status Line Format

*For any* successful response, THE output status line SHALL match the regex pattern `\[(HTTP|HTTPS)\] Status: \d{3} Time: \d+(\.\d+)?[mµn]?s`

**Validates: Requirements 2.7**

### Property 8: HTTP Method Setting

*For any* -X/--request flag value, THE Request_Builder SHALL set the request method to exactly that value (case-preserved).

**Validates: Requirements 3.1**

### Property 9: Header Accumulation

*For any* sequence of -H/--header flags, THE Request_Builder SHALL include ALL specified headers in the request, with later duplicates overwriting earlier ones for the same header name.

**Validates: Requirements 3.2, 8.4**

### Property 10: Data Flag Behavior

*For any* -d/--data flag value:
- THE Request_Builder SHALL set the request body to that value
- IF no -X flag is provided, THE method SHALL default to POST

**Validates: Requirements 3.3**

### Property 11: Raw Data Preservation

*For any* --data-raw flag value containing special characters (@, =, &), THE Request_Builder SHALL preserve those characters literally in the request body without interpretation.

**Validates: Requirements 3.4**

### Property 12: Basic Auth Header Generation

*For any* -u/--user flag value in format "user:password", THE Request_Builder SHALL set the Authorization header to "Basic " + base64(user:password).

**Validates: Requirements 3.5**

### Property 13: Cookie Header Setting

*For any* --cookie flag value, THE Request_Builder SHALL set the Cookie header to exactly that value.

**Validates: Requirements 3.6**

### Property 14: User-Agent Header Setting

*For any* --user-agent flag value, THE Request_Builder SHALL set the User-Agent header to exactly that value.

**Validates: Requirements 3.7**

### Property 15: Referer Header Setting

*For any* --referer flag value, THE Request_Builder SHALL set the Referer header to exactly that value.

**Validates: Requirements 3.8**

### Property 16: HEAD Request Flag

*For any* request with -I/--head flag, THE Request_Builder SHALL set the method to "HEAD".

**Validates: Requirements 4.3**

### Property 17: JSON Flag Headers

*For any* request with --json flag, THE Request_Builder SHALL set both:
- Content-Type: application/json
- Accept: application/json

**Validates: Requirements 4.4**

### Property 18: Timeout Duration Parsing

*For any* --timeout, --connect-timeout, or --max-time flag with a valid Go duration string, THE Transport SHALL parse it to the correct time.Duration value. --max-time SHALL be treated identically to --timeout.

**Validates: Requirements 6.1, 6.2, 6.3**

### Property 19: Flag Order Independence

*For any* valid set of flags and target, THE CLI_Parser SHALL produce identical Options regardless of the order in which flags and target appear in the argument list.

**Validates: Requirements 8.1**

### Property 20: Short and Long Flag Equivalence

*For any* flag that has both short (-v) and long (--verbose) forms, THE CLI_Parser SHALL produce identical Options for either form.

**Validates: Requirements 8.2**

### Property 21: Flag Value Format Equivalence

*For any* flag that accepts a value, THE CLI_Parser SHALL produce identical results for space-separated (-H "value") and equals-separated (-H="value") formats.

**Validates: Requirements 8.3**

### Property 22: Unknown Flag Rejection

*For any* flag not in the supported set, THE CLI_Parser SHALL return an error that maps to exit code 2.

**Validates: Requirements 8.5**

## Error Handling

### Error Classification

| Error Type | Exit Code | Description |
|------------|-----------|-------------|
| URLParseError | 3 | Invalid target format |
| NoRouteError | 6 | DNS resolution failed or no route |
| ConnectionError | 7 | Connection refused or reset |
| TimeoutError | 28 | Request or connect timeout |
| TLSError | 35 | Certificate validation or handshake failure |
| UnknownFlagError | 2 | Unrecognized CLI flag |

### Error Flow

1. CLI parsing errors → immediate exit with code 2
2. URL parsing errors → immediate exit with code 3
3. Network errors → classified and mapped to appropriate exit code
4. All errors print to stderr before exit

## Testing Strategy

### Property-Based Testing

Use `github.com/leanovate/gopter` for property-based testing in Go.

**Configuration:**
- Minimum 100 iterations per property test
- Each test tagged with property number and requirements reference

**Test Files:**
- `target_parser_test.go` - Properties 1-4
- `request_builder_test.go` - Properties 8-17
- `transport_test.go` - Properties 5-6, 18
- `cli_parser_test.go` - Properties 19-22
- `output_test.go` - Property 7

### Unit Tests

Unit tests complement property tests for:
- Specific edge cases (empty strings, special characters)
- Error message content verification
- Integration points between components

### Test Organization

```
curlz/
├── cmd/
│   └── curlz/
│       └── main.go
├── internal/
│   ├── cli/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── target/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── protocol/
│   │   ├── detector.go
│   │   └── detector_test.go
│   ├── request/
│   │   ├── builder.go
│   │   └── builder_test.go
│   ├── transport/
│   │   ├── transport.go
│   │   └── transport_test.go
│   └── output/
│       ├── handler.go
│       └── handler_test.go
├── go.mod
├── go.sum
└── Makefile
```
