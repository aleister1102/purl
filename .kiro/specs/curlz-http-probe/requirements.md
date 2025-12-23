# Requirements Document

## Introduction

Curlz is a drop-in curl replacement CLI tool written in Go that provides automatic HTTP/HTTPS protocol detection while preserving full curl compatibility. The tool simplifies working with IP:PORT targets by automatically probing protocols and falling back gracefully, making it ideal for security testing, API exploration, and network debugging.

## Glossary

- **Curlz**: The CLI tool that wraps HTTP client functionality with auto-protocol detection
- **Target**: The destination URL or IP:PORT combination to send requests to
- **Protocol_Detector**: The component responsible for determining whether HTTP or HTTPS should be used
- **Request_Builder**: The component that constructs HTTP requests from CLI flags
- **Transport**: The underlying HTTP transport layer handling connections and TLS
- **Auto_Mode**: Default behavior where Curlz tries HTTP first, then falls back to HTTPS

## Requirements

### Requirement 1: Target Parsing and URL Construction

**User Story:** As a user, I want to specify targets in flexible formats (IP:PORT, host:PORT/path, or full URLs), so that I can quickly probe endpoints without typing full URLs.

#### Acceptance Criteria

1. WHEN a user provides a target in format "IP:PORT/path", THE Target_Parser SHALL construct a valid URL with the detected or specified protocol
2. WHEN a user provides a target in format "host:PORT", THE Target_Parser SHALL append the default path "/" if no path is specified
3. WHEN a user provides a full URL with protocol (http:// or https://), THE Target_Parser SHALL use the specified protocol without auto-detection
4. IF a target cannot be parsed into a valid URL, THEN THE Target_Parser SHALL return a descriptive error message and exit with code 3

### Requirement 2: Auto Protocol Detection

**User Story:** As a user, I want Curlz to automatically detect whether a target uses HTTP or HTTPS, so that I don't need to manually test both protocols.

#### Acceptance Criteria

1. WHEN --auto mode is active (default) and no protocol is specified, THE Protocol_Detector SHALL first attempt HTTP connection with a 3-second timeout
2. WHEN HTTP attempt succeeds with status 200-399, THE Protocol_Detector SHALL use HTTP and return the response
3. WHEN HTTP attempt fails or returns non-success status, THE Protocol_Detector SHALL fall back to HTTPS with a 7-second timeout
4. WHEN connecting to IP addresses via HTTPS in auto mode, THE Transport SHALL use InsecureSkipVerify by default
5. WHEN --proto http is specified, THE Protocol_Detector SHALL only attempt HTTP without fallback
6. WHEN --proto https is specified, THE Protocol_Detector SHALL only attempt HTTPS without fallback
7. THE Curlz SHALL print protocol detection result in format "[PROTO] Status: CODE Time: Xs" before response body

### Requirement 3: Curl-Compatible Request Flags

**User Story:** As a user familiar with curl, I want to use the same flags I know from curl, so that I can switch tools without learning new syntax.

#### Acceptance Criteria

1. WHEN -X or --request flag is provided, THE Request_Builder SHALL set the HTTP method accordingly
2. WHEN -H or --header flag is provided (can be repeated), THE Request_Builder SHALL add each header to the request
3. WHEN -d or --data flag is provided, THE Request_Builder SHALL set the request body and default method to POST
4. WHEN --data-raw flag is provided, THE Request_Builder SHALL set the request body without special character interpretation
5. WHEN -u or --user flag is provided in format "user:password", THE Request_Builder SHALL set Basic Authentication header
6. WHEN --cookie flag is provided, THE Request_Builder SHALL add the Cookie header
7. WHEN --user-agent flag is provided, THE Request_Builder SHALL set the User-Agent header
8. WHEN --referer flag is provided, THE Request_Builder SHALL set the Referer header

### Requirement 4: Output Control Flags

**User Story:** As a user, I want to control how responses are displayed and saved, so that I can integrate Curlz into scripts and workflows.

#### Acceptance Criteria

1. WHEN -v or --verbose flag is provided, THE Curlz SHALL print request details and response headers to stderr
2. WHEN -o or --output flag is provided with a filename, THE Curlz SHALL write response body to the specified file
3. WHEN -I or --head flag is provided, THE Curlz SHALL send HEAD request and display only response headers
4. WHEN --json flag is provided, THE Curlz SHALL set Content-Type and Accept headers to application/json
5. WHEN --verbose-tls flag is provided, THE Curlz SHALL print TLS handshake details to stderr
6. THE Curlz SHALL stream response body to stdout by default

### Requirement 5: TLS and Security Options

**User Story:** As a security tester, I want fine-grained control over TLS verification, so that I can test endpoints with self-signed certificates or enforce strict validation.

#### Acceptance Criteria

1. WHEN -k or --insecure flag is provided, THE Transport SHALL skip TLS certificate verification
2. WHEN --cacert flag is provided with a CA certificate path, THE Transport SHALL use the specified CA for verification
3. WHEN --cert flag is provided with a client certificate path, THE Transport SHALL use the certificate for client authentication
4. WHEN --key flag is provided with a private key path, THE Transport SHALL use the key for client certificate authentication
5. WHEN --strict-ssl flag is provided, THE Transport SHALL enforce full certificate chain validation even for IP targets
6. WHEN connecting to IP addresses without --strict-ssl, THE Transport SHALL default to InsecureSkipVerify

### Requirement 6: Timeout and Connection Control

**User Story:** As a user, I want to control connection timeouts, so that I can handle slow or unresponsive endpoints gracefully.

#### Acceptance Criteria

1. WHEN --timeout flag is provided with a duration (e.g., "10s"), THE Transport SHALL use that duration as the total request timeout
2. WHEN --connect-timeout flag is provided, THE Transport SHALL use that duration for the connection phase only
3. WHEN --max-time flag is provided (curl alias), THE Transport SHALL treat it as equivalent to --timeout
4. THE Curlz SHALL use a default timeout of 10 seconds when no timeout is specified
5. IF a request times out, THEN THE Curlz SHALL exit with code 28

### Requirement 7: Exit Codes

**User Story:** As a script author, I want Curlz to return meaningful exit codes, so that I can handle errors programmatically.

#### Acceptance Criteria

1. WHEN a request completes successfully, THE Curlz SHALL exit with code 0
2. WHEN URL parsing fails, THE Curlz SHALL exit with code 3
3. WHEN no route to host exists, THE Curlz SHALL exit with code 6
4. WHEN connection is refused or fails, THE Curlz SHALL exit with code 7
5. WHEN request times out, THE Curlz SHALL exit with code 28
6. WHEN TLS/SSL error occurs, THE Curlz SHALL exit with code 35

### Requirement 8: CLI Argument Parsing

**User Story:** As a user, I want flexible argument ordering like curl, so that I can place flags before or after the target URL.

#### Acceptance Criteria

1. THE CLI_Parser SHALL accept flags in any order relative to the target
2. THE CLI_Parser SHALL support both short (-v) and long (--verbose) flag formats
3. THE CLI_Parser SHALL support flag values with space (-H "Header: Value") or equals (-H="Header: Value")
4. WHEN multiple -H flags are provided, THE CLI_Parser SHALL collect all headers
5. IF an unknown flag is provided, THEN THE CLI_Parser SHALL print an error and exit with code 2
