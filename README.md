# vanity-ssh-keygen

```
Generate a vanity SSH key that matches the input argument

Usage:
  vanity-ssh-keygen [match-string] [flags]

Flags:
      --config string       Config file
      --enable-metrics      Enable metrics server. (default true)
  -h, --help                help for vanity-ssh-keygen
  -t, --key-type string     Key type to generate (default "ed25519")
      --matcher string      Matcher used to find a vanity SSH key (default "ignorecase")
      --metrics-port int    Listening port for metrics server. (default 9101)
  -o, --output string       Output format. One of: pem-files|json-file. (default "pem-files")
      --output-dir string   Output directory. (default "./")
      --profile             Write pprof CPU profile to ./pprof
  -j, --threads int         Execution threads. Defaults to the number of logical CPU cores (default 8)
  -v, --version             version for vanity-ssh-keygen
```
