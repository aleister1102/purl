package errors

import (
	"fmt"
	"time"
)

// Exit code constants matching curl's exit codes
const (
	ExitSuccess       = 0
	ExitUnknownFlag   = 2
	ExitURLParse      = 3
	ExitNoRoute       = 6
	ExitConnectFailed = 7
	ExitTimeout       = 28
	ExitTLSError      = 35
)

// URLParseError represents an error parsing the target URL
type URLParseError struct {
	Input   string
	Message string
}

func (e *URLParseError) Error() string {
	return fmt.Sprintf("URL parse error: %s (input: %s)", e.Message, e.Input)
}

// ConnectionError represents a connection failure
type ConnectionError struct {
	Host  string
	Port  string
	Cause error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error to %s:%s: %v", e.Host, e.Port, e.Cause)
}

// TimeoutError represents a timeout during request or connection
type TimeoutError struct {
	Duration time.Duration
	Phase    string // "connect" or "request"
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout during %s phase: %v", e.Phase, e.Duration)
}

// TLSError represents a TLS/SSL error
type TLSError struct {
	Host  string
	Cause error
}

func (e *TLSError) Error() string {
	return fmt.Sprintf("TLS error for %s: %v", e.Host, e.Cause)
}

// UnknownFlagError represents an unknown CLI flag
type UnknownFlagError struct {
	Flag string
}

func (e *UnknownFlagError) Error() string {
	return fmt.Sprintf("unknown flag: %s", e.Flag)
}

// NoRouteError represents a DNS resolution or routing failure
type NoRouteError struct {
	Host  string
	Cause error
}

func (e *NoRouteError) Error() string {
	return fmt.Sprintf("no route to host %s: %v", e.Host, e.Cause)
}

// MapErrorToExitCode maps error types to curl-compatible exit codes
func MapErrorToExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	switch err.(type) {
	case *URLParseError:
		return ExitURLParse
	case *UnknownFlagError:
		return ExitUnknownFlag
	case *NoRouteError:
		return ExitNoRoute
	case *ConnectionError:
		return ExitConnectFailed
	case *TimeoutError:
		return ExitTimeout
	case *TLSError:
		return ExitTLSError
	default:
		// Default to connection error for unknown errors
		return ExitConnectFailed
	}
}
