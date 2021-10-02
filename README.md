# vanity-ssh-keygen

```
Usage: vanity-ssh-keygen <match-string>

Arguments:
  <match-string>

Flags:
  -h, --help                     Show context-sensitive help.
      --version                  Print version and exit
      --matcher="ignorecase"     Matcher used to find a vanity SSH key. One of: ignorecase,ignorecase-ed25519
  -t, --key-type="ed25519"       Key type to generate. One of: ed25519,rsa-2048,rsa-4096
  -j, --threads=8                Execution threads. Defaults to the number of logical CPU cores
      --profile                  Profile the process. Write pprof CPU profile to ./pprof
      --[no-]metrics             Enable metrics server.
      --metrics-port=9101        Listening port for metrics server.
  -o, --output="pem-files"       Output format. One of: pem-files|json-file.
      --output-dir="./"          Output directory.
      --stats-log-interval=2s    Statistics will be printed at this interval, set to 0 to disable
```
