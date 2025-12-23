package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/user/curlz/internal/errors"
)

// ParseArgs parses command-line arguments and returns Options
func ParseArgs(args []string) (*Options, error) {
	opts := &Options{
		Proto:   "auto",
		Timeout: 10 * time.Second,
	}

	app := &cli.App{
		Name:  "curlz",
		Usage: "curl-compatible HTTP probe with auto protocol detection",
		Flags: buildFlags(),
		Action: func(c *cli.Context) error {
			// Extract target from positional arguments
			if c.NArg() == 0 {
				return fmt.Errorf("target URL required")
			}
			opts.Target = c.Args().Get(0)

			// Parse all flags into options
			return parseFlags(c, opts)
		},
		HelpName: "curlz",
	}

	// Parse arguments - urfave/cli will handle flag parsing
	err := app.Run(args)
	if err != nil {
		// Check if it's a flag-related error
		if strings.Contains(err.Error(), "flag provided but not defined") ||
			strings.Contains(err.Error(), "unknown flag") {
			return nil, &errors.UnknownFlagError{Flag: extractFlagName(err.Error())}
		}
		return nil, err
	}

	return opts, nil
}

// buildFlags creates all CLI flags matching curl compatibility
func buildFlags() []cli.Flag {
	return []cli.Flag{
		// Request method
		&cli.StringFlag{
			Name:    "request",
			Aliases: []string{"X"},
			Usage:   "Specify request method (GET, POST, etc.)",
		},

		// Headers
		&cli.StringSliceFlag{
			Name:    "header",
			Aliases: []string{"H"},
			Usage:   "Pass custom header(s) to server (can be repeated)",
		},

		// Data/Body
		&cli.StringFlag{
			Name:    "data",
			Aliases: []string{"d"},
			Usage:   "HTTP POST data",
		},
		&cli.StringFlag{
			Name:  "data-raw",
			Usage: "HTTP POST data without special character interpretation",
		},

		// Authentication
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "Server user and password (user:password)",
		},

		// Cookies and headers
		&cli.StringFlag{
			Name:  "cookie",
			Usage: "Send cookie(s) to server",
		},
		&cli.StringFlag{
			Name:  "user-agent",
			Usage: "Send User-Agent string to server",
		},
		&cli.StringFlag{
			Name:  "referer",
			Usage: "Send Referer header to server",
		},

		// Output control
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Make the operation more talkative",
		},
		&cli.BoolFlag{
			Name:  "verbose-tls",
			Usage: "Print TLS handshake details",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Write output to file instead of stdout",
		},
		&cli.BoolFlag{
			Name:    "head",
			Aliases: []string{"I"},
			Usage:   "HEAD request",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "Set Content-Type and Accept to application/json",
		},

		// TLS/SSL
		&cli.BoolFlag{
			Name:    "insecure",
			Aliases: []string{"k"},
			Usage:   "Allow insecure server connections when using SSL",
		},
		&cli.StringFlag{
			Name:  "cacert",
			Usage: "CA certificate to verify peer against",
		},
		&cli.StringFlag{
			Name:  "cert",
			Usage: "Client certificate file",
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "Private key file",
		},
		&cli.BoolFlag{
			Name:  "strict-ssl",
			Usage: "Enforce strict SSL certificate validation",
		},

		// Protocol
		&cli.StringFlag{
			Name:  "proto",
			Usage: "Protocol to use (auto, http, https)",
			Value: "auto",
		},

		// Timeouts
		&cli.StringFlag{
			Name:  "timeout",
			Usage: "Maximum time allowed for the operation (e.g., 10s, 1m)",
		},
		&cli.StringFlag{
			Name:  "connect-timeout",
			Usage: "Maximum time allowed for connection",
		},
		&cli.StringFlag{
			Name:  "max-time",
			Usage: "Maximum time allowed for the operation (alias for --timeout)",
		},
	}
}

// parseFlags extracts flag values from cli.Context and populates Options
func parseFlags(c *cli.Context, opts *Options) error {
	// Request method
	if c.IsSet("request") {
		opts.Method = c.String("request")
	}

	// Headers
	if c.IsSet("header") {
		opts.Headers = c.StringSlice("header")
	}

	// Data/Body
	if c.IsSet("data") {
		opts.Data = c.String("data")
	}
	if c.IsSet("data-raw") {
		opts.DataRaw = c.String("data-raw")
	}

	// Authentication
	if c.IsSet("user") {
		opts.User = c.String("user")
	}

	// Cookies and headers
	if c.IsSet("cookie") {
		opts.Cookie = c.String("cookie")
	}
	if c.IsSet("user-agent") {
		opts.UserAgent = c.String("user-agent")
	}
	if c.IsSet("referer") {
		opts.Referer = c.String("referer")
	}

	// Output control
	if c.IsSet("verbose") {
		opts.Verbose = c.Bool("verbose")
	}
	if c.IsSet("verbose-tls") {
		opts.VerboseTLS = c.Bool("verbose-tls")
	}
	if c.IsSet("output") {
		opts.Output = c.String("output")
	}
	if c.IsSet("head") {
		opts.Head = c.Bool("head")
	}
	if c.IsSet("json") {
		opts.JSON = c.Bool("json")
	}

	// TLS/SSL
	if c.IsSet("insecure") {
		opts.Insecure = c.Bool("insecure")
	}
	if c.IsSet("cacert") {
		opts.CACert = c.String("cacert")
	}
	if c.IsSet("cert") {
		opts.Cert = c.String("cert")
	}
	if c.IsSet("key") {
		opts.Key = c.String("key")
	}
	if c.IsSet("strict-ssl") {
		opts.StrictSSL = c.Bool("strict-ssl")
	}

	// Protocol
	if c.IsSet("proto") {
		proto := c.String("proto")
		if proto != "auto" && proto != "http" && proto != "https" {
			return fmt.Errorf("invalid protocol: %s (must be auto, http, or https)", proto)
		}
		opts.Proto = proto
	}

	// Timeouts
	if c.IsSet("timeout") {
		duration, err := time.ParseDuration(c.String("timeout"))
		if err != nil {
			return fmt.Errorf("invalid timeout format: %v", err)
		}
		opts.Timeout = duration
	}

	if c.IsSet("connect-timeout") {
		duration, err := time.ParseDuration(c.String("connect-timeout"))
		if err != nil {
			return fmt.Errorf("invalid connect-timeout format: %v", err)
		}
		opts.ConnectTimeout = duration
	}

	// max-time is an alias for timeout
	if c.IsSet("max-time") {
		duration, err := time.ParseDuration(c.String("max-time"))
		if err != nil {
			return fmt.Errorf("invalid max-time format: %v", err)
		}
		opts.Timeout = duration
	}

	return nil
}

// extractFlagName extracts the flag name from an error message
func extractFlagName(errMsg string) string {
	// Try to extract flag name from common error patterns
	if idx := strings.Index(errMsg, "flag provided but not defined: -"); idx != -1 {
		start := idx + len("flag provided but not defined: ")
		end := strings.IndexAny(errMsg[start:], " \n")
		if end == -1 {
			return errMsg[start:]
		}
		return errMsg[start : start+end]
	}
	return "unknown"
}
