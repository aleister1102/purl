package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/user/purl/internal/cli"
	"github.com/user/purl/internal/errors"
	"github.com/user/purl/internal/output"
	"github.com/user/purl/internal/protocol"
	"github.com/user/purl/internal/request"
	"github.com/user/purl/internal/target"
	"github.com/user/purl/internal/transport"
)

func main() {
	// Parse CLI arguments (pass full args including program name for urfave/cli)
	opts, err := cli.ParseArgs(os.Args)
	if err != nil {
		printError(err)
		os.Exit(errors.MapErrorToExitCode(err))
	}

	// Execute the main workflow
	exitCode := run(opts)
	os.Exit(exitCode)
}

// run executes the main workflow: parse target → detect protocol → build request → execute → output
func run(opts *cli.Options) int {
	// Step 1: Parse target URL
	parsedTarget, err := target.ParseTarget(opts.Target)
	if err != nil {
		printError(err)
		return errors.MapErrorToExitCode(err)
	}

	// Step 2: Detect protocol (auto or manual)
	probeResult, err := protocol.DetectProtocol(parsedTarget, opts)
	if err != nil {
		printError(err)
		return errors.MapErrorToExitCode(err)
	}

	// If protocol detection failed, return the error
	if probeResult.Error != nil {
		printError(probeResult.Error)
		return errors.MapErrorToExitCode(probeResult.Error)
	}

	// Step 3: Update the parsed target URL with the detected protocol
	parsedTarget.URL.Scheme = probeResult.Protocol

	// Step 4: Build the actual request (not just the probe)
	ctx, cancel := context.WithTimeout(context.Background(), transport.ApplyTimeouts(opts))
	defer cancel()

	req, err := request.BuildRequest(ctx, parsedTarget, opts)
	if err != nil {
		printError(err)
		return errors.MapErrorToExitCode(err)
	}

	// Step 5: Create transport and execute the request
	tr, err := transport.NewTransport(opts, parsedTarget)
	if err != nil {
		printError(err)
		return errors.MapErrorToExitCode(err)
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   transport.ApplyTimeouts(opts),
	}

	resp, err := client.Do(req)
	if err != nil {
		printError(err)
		return errors.MapErrorToExitCode(err)
	}
	defer resp.Body.Close()

	// Update probe result with actual response
	probeResult.Response = resp
	probeResult.StatusCode = resp.StatusCode

	// Step 6: Output the response
	handler := output.NewHandler(opts)
	if err := handler.WriteResponse(req, probeResult); err != nil {
		printError(err)
		return errors.ExitConnectFailed
	}

	return errors.ExitSuccess
}

// printError prints an error message to stderr
func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "purl: %v\n", err)
	}
}
