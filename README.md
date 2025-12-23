# purl (Probing cURL)

A curl-compatible HTTP probe with automatic protocol detection. Purl simplifies working with IP:PORT targets by automatically probing HTTP/HTTPS protocols and falling back gracefully.

## Features

- **Auto Protocol Detection**: Automatically tries HTTP first (3s timeout), then falls back to HTTPS (7s timeout)
- **Curl-Compatible**: Supports familiar curl flags and syntax
- **Flexible Target Formats**: Accept IP:PORT, host:PORT/path, or full URLs
- **Smart TLS Handling**: Automatically skips certificate verification for IP addresses (configurable)

## Installation

```bash
go install github.com/aleister1102/purl/cmd/purl@latest
```

Or build from source:

```bash
git clone https://github.com/aleister1102/purl.git
cd purl
go build -o purl ./cmd/purl
```

## Quick Start

```bash
# Simple GET request with auto protocol detection
purl localhost:8080

# Specify protocol explicitly
purl --proto https example.com

# POST with JSON data
purl -X POST --json -d '{"key":"value"}' api.example.com/endpoint

# Verbose output
purl -v https://example.com

# Custom headers
purl -H "Authorization: Bearer token" -H "Accept: application/json" api.example.com
```

## Usage

```
purl [options] <target>
```

### Target Formats

- `IP:PORT` - e.g., `192.168.1.1:8080`
- `IP:PORT/path` - e.g., `192.168.1.1:8080/api/v1`
- `host:PORT` - e.g., `localhost:3000`
- `host:PORT/path` - e.g., `api.example.com:443/users`
- Full URL - e.g., `https://example.com/path`

### Common Options

#### Request Options
- `-X, --request <method>` - HTTP method (GET, POST, PUT, DELETE, etc.)
- `-H, --header <header>` - Add custom header (can be repeated)
- `-d, --data <data>` - HTTP POST data
- `--data-raw <data>` - POST data without special character interpretation
- `-u, --user <user:pass>` - Basic authentication

#### Output Options
- `-v, --verbose` - Verbose output (request details to stderr)
- `-o, --output <file>` - Write response to file
- `-I, --head` - Send HEAD request
- `--json` - Set Content-Type and Accept to application/json

#### TLS/Security Options
- `-k, --insecure` - Skip TLS certificate verification
- `--strict-ssl` - Enforce strict SSL validation (even for IP addresses)
- `--cacert <file>` - CA certificate for verification
- `--cert <file>` - Client certificate
- `--key <file>` - Client private key

#### Protocol Options
- `--proto <protocol>` - Force protocol: `auto` (default), `http`, or `https`

#### Timeout Options
- `--timeout <duration>` - Maximum time for operation (e.g., `10s`, `1m`)
- `--connect-timeout <duration>` - Connection timeout
- `--max-time <duration>` - Alias for --timeout

## Examples

### Auto Protocol Detection

```bash
# Tries HTTP first, falls back to HTTPS
purl 192.168.1.1:8080
```

Output:
```
[HTTP] Status: 200 Time: 45ms
{"status": "ok"}
```

### POST Request with JSON

```bash
purl -X POST --json -d '{"username":"admin","password":"secret"}' localhost:3000/login
```

### Custom Headers and Authentication

```bash
purl -H "X-API-Key: secret" -u admin:password api.example.com/users
```

### Verbose Mode

```bash
purl -v https://example.com
```

Output includes request details:
```
> GET / HTTP/1.1
> Host: example.com
> User-Agent: purl
> Accept: */*
>
[HTTPS] Status: 200 Time: 123ms
<!DOCTYPE html>...
```

### Save Response to File

```bash
purl -o response.json https://api.example.com/data
```

### Skip Certificate Verification

```bash
# For self-signed certificates
purl -k https://192.168.1.1:8443
```

## Exit Codes

- `0` - Success
- `2` - Unknown flag
- `3` - URL parse error
- `6` - No route to host
- `7` - Connection failed
- `28` - Timeout
- `35` - TLS/SSL error

## Differences from curl

While purl aims for curl compatibility, there are some key differences:

1. **Auto Protocol Detection**: By default, purl tries HTTP first, then HTTPS
2. **IP Address TLS**: Automatically skips certificate verification for IP addresses (unless `--strict-ssl` is used)
3. **Simplified Output**: Status line format is `[PROTO] Status: CODE Time: Xs`

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run property-based tests
go test -v ./internal/target
```

### Building

```bash
go build -o purl ./cmd/purl
```

## License

MIT License - see LICENSE file for details
