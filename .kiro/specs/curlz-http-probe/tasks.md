# Implementation Plan: Curlz HTTP Probe

## Overview

This plan implements curlz as a modular Go CLI tool with curl-compatible flags and automatic HTTP/HTTPS protocol detection. Tasks are ordered to build foundational components first, then wire them together incrementally.

## Tasks

- [x] 1. Project setup and core types
  - Initialize Go module with `go mod init github.com/user/curlz`
  - Add dependency: `github.com/urfave/cli/v2`
  - Add test dependency: `github.com/leanovate/gopter`
  - Create directory structure: `cmd/curlz/`, `internal/cli/`, `internal/target/`, `internal/protocol/`, `internal/request/`, `internal/transport/`, `internal/output/`
  - Define `Options` struct in `internal/cli/options.go`
  - Define error types and exit codes in `internal/errors/errors.go`
  - _Requirements: 7.1-7.6_

- [-] 2. Target parser implementation
  - [x] 2.1 Implement ParseTarget function
    - Parse IP:PORT/path, host:PORT, and full URL formats
    - Detect if target is IP address vs hostname
    - Track whether protocol was explicit
    - Append default "/" path when missing
    - Return URLParseError for invalid inputs
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [x]* 2.2 Write property test for URL construction round-trip
    - **Property 1: URL Construction Round-Trip**
    - **Validates: Requirements 1.1, 1.2, 1.3**
  - [x]* 2.3 Write property test for default path appending
    - **Property 2: Default Path Appending**
    - **Validates: Requirements 1.2**
  - [x]* 2.4 Write property test for explicit protocol preservation
    - **Property 3: Explicit Protocol Preservation**
    - **Validates: Requirements 1.3**
  - [x]* 2.5 Write property test for invalid target error mapping
    - **Property 4: Invalid Target Error Mapping**
    - **Validates: Requirements 1.4, 7.2**

- [x] 3. Checkpoint - Target parser complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Request builder implementation
  - [x] 4.1 Implement BuildRequest function
    - Create http.Request with method, URL, headers, body
    - Handle -X method override
    - Handle -d/--data body with POST default
    - Handle --data-raw without interpretation
    - _Requirements: 3.1, 3.2, 3.3, 3.4_
  - [x] 4.2 Implement authentication and header helpers
    - Basic auth from -u flag (base64 encoding)
    - Cookie header from --cookie flag
    - User-Agent from --user-agent flag
    - Referer from --referer flag
    - _Requirements: 3.5, 3.6, 3.7, 3.8_
  - [x] 4.3 Implement special flags
    - HEAD method for -I/--head flag
    - JSON headers for --json flag
    - _Requirements: 4.3, 4.4_
  - [x]* 4.4 Write property test for HTTP method setting
    - **Property 8: HTTP Method Setting**
    - **Validates: Requirements 3.1**
  - [x]* 4.5 Write property test for header accumulation
    - **Property 9: Header Accumulation**
    - **Validates: Requirements 3.2, 8.4**
  - [x]* 4.6 Write property test for data flag behavior
    - **Property 10: Data Flag Behavior**
    - **Validates: Requirements 3.3**
  - [x]* 4.7 Write property test for raw data preservation
    - **Property 11: Raw Data Preservation**
    - **Validates: Requirements 3.4**
  - [x]* 4.8 Write property test for basic auth header generation
    - **Property 12: Basic Auth Header Generation**
    - **Validates: Requirements 3.5**
  - [x]* 4.9 Write property tests for cookie, user-agent, referer headers
    - **Property 13: Cookie Header Setting**
    - **Property 14: User-Agent Header Setting**
    - **Property 15: Referer Header Setting**
    - **Validates: Requirements 3.6, 3.7, 3.8**
  - [x]* 4.10 Write property test for HEAD request flag
    - **Property 16: HEAD Request Flag**
    - **Validates: Requirements 4.3**
  - [x]* 4.11 Write property test for JSON flag headers
    - **Property 17: JSON Flag Headers**
    - **Validates: Requirements 4.4**

- [x] 5. Checkpoint - Request builder complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Transport implementation
  - [x] 6.1 Implement NewTransport function
    - Configure http.Transport with TLS settings
    - Handle InsecureSkipVerify based on target type and flags
    - Support --strict-ssl override
    - _Requirements: 5.1, 5.5, 5.6_
  - [x] 6.2 Implement timeout handling
    - Parse duration strings for --timeout, --connect-timeout, --max-time
    - Apply context-based timeouts
    - Default to 10s when not specified
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  - [x]* 6.3 Write property test for TLS InsecureSkipVerify configuration
    - **Property 6: TLS InsecureSkipVerify Configuration**
    - **Validates: Requirements 2.4, 5.1, 5.5, 5.6**
  - [x]* 6.4 Write property test for timeout duration parsing
    - **Property 18: Timeout Duration Parsing**
    - **Validates: Requirements 6.1, 6.2, 6.3**

- [x] 7. Checkpoint - Transport complete
  - Ensure all tests pass, ask the user if questions arise.

- [-] 8. Protocol detector implementation
  - [x] 8.1 Implement DetectProtocol function
    - Auto mode: HTTP first (3s timeout), HTTPS fallback (7s timeout)
    - Manual mode: use --proto value directly
    - Return ProbeResult with protocol, status, duration
    - _Requirements: 2.1, 2.2, 2.3, 2.5, 2.6_
  - [ ]* 8.2 Write property test for protocol override behavior
    - **Property 5: Protocol Override Behavior**
    - **Validates: Requirements 2.5, 2.6**

- [x] 9. Output handler implementation
  - [x] 9.1 Implement output formatting
    - Format status line: "[PROTO] Status: CODE Time: Xs"
    - Stream response body to stdout
    - Write to file for -o/--output flag
    - _Requirements: 2.7, 4.2, 4.6_
  - [x] 9.2 Implement verbose output
    - Print request details to stderr for -v flag
    - Print TLS details for --verbose-tls flag
    - _Requirements: 4.1, 4.5_
  - [ ]* 9.3 Write property test for status line format
    - **Property 7: Status Line Format**
    - **Validates: Requirements 2.7**

- [x] 10. CLI parser implementation
  - [x] 10.1 Implement CLI with urfave/cli
    - Define all flags matching curl compatibility
    - Support short and long flag forms
    - Support space and equals value formats
    - Collect multiple -H flags into slice
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  - [x] 10.2 Implement flag validation
    - Reject unknown flags with exit code 2
    - Validate flag value formats
    - _Requirements: 8.5_
  - [x]* 10.3 Write property test for flag order independence
    - **Property 19: Flag Order Independence**
    - **Validates: Requirements 8.1**
  - [x]* 10.4 Write property test for short/long flag equivalence
    - **Property 20: Short and Long Flag Equivalence**
    - **Validates: Requirements 8.2**
  - [x]* 10.5 Write property test for flag value format equivalence
    - **Property 21: Flag Value Format Equivalence**
    - **Validates: Requirements 8.3**
  - [x]* 10.6 Write property test for unknown flag rejection
    - **Property 22: Unknown Flag Rejection**
    - **Validates: Requirements 8.5**

- [x] 11. Checkpoint - CLI parser complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 12. Main entry point and wiring
  - [x] 12.1 Implement main.go
    - Wire CLI parser → Target parser → Protocol detector → Request builder → Transport → Output
    - Handle errors and map to exit codes
    - _Requirements: 7.1-7.6_
  - [x] 12.2 Implement error-to-exit-code mapping
    - Map error types to curl-compatible exit codes
    - Print error messages to stderr
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [x] 13. Final checkpoint
  - Ensure all tests pass, ask the user if questions arise.
  - Verify `go build` produces working binary
  - Test basic usage: `./curlz localhost:8080`

## Notes

- Tasks marked with `*` are optional property-based tests that can be skipped for faster MVP
- Each property test should run minimum 100 iterations
- Use `github.com/leanovate/gopter` for property-based testing
- All tests should be tagged with property number for traceability
