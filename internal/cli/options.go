package cli

import "time"

// Options holds all parsed CLI flags and target information
type Options struct {
	// Target
	Target string
	Proto  string // "auto", "http", "https"

	// Request
	Method   string
	Headers  []string
	Data     string
	DataRaw  string
	User     string // user:password
	Cookie   string
	UserAgent string
	Referer  string

	// Output
	Verbose    bool
	VerboseTLS bool
	Output     string
	Head       bool
	JSON       bool

	// TLS
	Insecure  bool
	CACert    string
	Cert      string
	Key       string
	StrictSSL bool

	// Timeouts
	Timeout        time.Duration
	ConnectTimeout time.Duration
}
