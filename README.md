# vanity-ssh-keygen

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

