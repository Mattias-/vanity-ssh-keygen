# vanity-ssh-keygen

A high-performance SSH vanity key generator written in Go. It allows you to find SSH public keys that contain specific strings or patterns.

## Features

- **Multi-threaded:** Automatically utilizes all available CPU cores.
- **High Performance:** Optimized ED25519 generation and matching (zero-allocation hot loop).
- **Multiple Algorithms:** Supports ED25519 and RSA (2048/4096 bit).
- **Flexible Matching:** Support for case-insensitive matching.
- **Graceful Shutdown:** Handles `SIGINT` and `SIGTERM` to stop workers cleanly.
- **Observability:** Built-in support for OpenTelemetry metrics, logs, and profiling (pprof/Pyroscope).

## Usage

### Examples

Find an ED25519 key containing "abc" (case-insensitive):
```bash
./vanity-ssh-keygen abc
```

Find an RSA-4096 key:
```bash
./vanity-ssh-keygen test -t rsa-4096
```

Use specific number of threads and output to JSON:
```bash
./vanity-ssh-keygen supersecret -j 4 -o json-file
```

### Full CLI Usage

<!-- vanity-ssh-keygen-usage:start -->
```text
Usage: vanity-ssh-keygen <match-string> [flags]

Arguments:
  <match-string>

Flags:
  -h, --help                     Show context-sensitive help.
      --version                  Print version and exit
      --debug                    Enable debug logging
      --matcher="ignorecase"     Matcher used to find a vanity SSH key. One of:
                                 ignorecase,ignorecase-ed25519
  -t, --key-type="ed25519"       Key type to generate. One of:
                                 ed25519,rsa-2048,rsa-4096
  -j, --threads=8                Execution threads. Defaults to the number of
                                 logical CPU cores
      --profile                  Profile the process. Write pprof CPU profile to
                                 ./pprof
      --pyroscope-profile        Profile the process and upload data to
                                 Pyroscope
      --metrics                  Enable metrics server.
      --otel-logs                Enable otel logs.
  -o, --output="pem-files"       Output format. One of: pem-files|json-file.
      --output-dir="./"          Output directory.
      --stats-log-interval=2s    Statistics will be printed at this interval,
                                 set to 0 to disable
```
<!-- vanity-ssh-keygen-usage:end -->

## Performance

The tool is highly optimized for ED25519. The key generation loop avoids memory allocations for public key serialization and matching, maximizing throughput.

### Estimated Search Times

The following table shows the estimated time to find an **ED25519** key with a specific length match string using the `ignorecase-ed25519` matcher.

*Based on an Apple M2 (8 cores) reaching ~330,000 keys/second:*

| Match Length | Avg. Keys to Test | Avg. Time (1 Thread) | Avg. Time (8 Threads) |
|--------------|-------------------|----------------------|-----------------------|
| 1 char       | < 5               | < 0.01s              | < 0.01s               |
| 2 chars      | ~25               | < 0.01s              | < 0.01s               |
| 3 chars      | ~850              | 0.02s                | < 0.01s               |
| 4 chars      | ~27,000           | 0.6s                 | 0.1s                  |
| 5 chars      | ~860,000          | 20s                  | 2.5s                  |
| 6 chars      | ~27,000,000       | 11m                  | 1.5m                  |
| 7 chars      | ~880,000,000      | 6h                   | 45m                   |
| 8 chars      | ~28,000,000,000   | 8 days               | 1 day                 |
| 9 chars      | ~900,000,000,000  | 8 months             | 1 month               |

*Note: Since the matcher searches anywhere within the 43-character Base64 public key, the probability of finding a match is significantly higher than if it were restricted to the prefix. Each character in a case-insensitive search has a ~1/32 probability of matching a letter.*

### Benchmarking

To run the internal performance benchmarks:

```bash
go test -bench . -benchmem ./...
```

## Development

### Requirements
- Go 1.22+
- `golangci-lint` (for linting)

### Common Tasks
- `make build`: Build the binary.
- `make test`: Run unit tests.
- `make lint-fix`: Automatically fix linting and formatting issues.
- `make smoke-test`: Run a quick end-to-end test.
