package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/google/uuid"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/worker"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/workerpool"
)

const (
	serviceName = "vanity-ssh-keygen"
)

var (
	version    = "dev"
	instanceID = uuid.NewString()
)

type Metadata struct {
	FindString string `json:"findstring"`
	Time       int64  `json:"time"`
}

type OutputData struct {
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_key"`
	Metadata   Metadata `json:"metadata"`
}

type cli struct {
	Version          kong.VersionFlag `help:"Print version and exit"`
	MatchString      string           `arg:""`
	Matcher          string           `help:"Matcher used to find a vanity SSH key. One of: ${matchers}" default:"${default_matcher}" enum:"${matchers}"`
	KeyType          string           `short:"t" help:"Key type to generate. One of: ${keytypes}" enum:"${keytypes}" default:"${default_keytype}"`
	Threads          int              `short:"j" help:"Execution threads. Defaults to the number of logical CPU cores" default:"${default_threads}"`
	Profile          bool             `help:"Profile the process. Write pprof CPU profile to ./pprof" default:"false"`
	PyroscopeProfile bool             `help:"Profile the process and upload data to Pyroscope" default:"false"`
	Metrics          bool             `help:"Enable metrics server." default:"false"`
	Output           string           `short:"o" help:"Output format. One of: pem-files|json-file." default:"pem-files"`
	OutputDir        string           `help:"Output directory." default:"./" type:"existingdir"`
	StatsLogInterval time.Duration    `help:"Statistics will be printed at this interval, set to 0 to disable" default:"2s"`
}

func runKeygen(c cli, matcher matcher.Matcher, kg keygen.Keygen) {
	var workers []workerpool.Worker[chan keygen.SSHKey]
	for i := 0; i < c.Threads; i++ {
		w := &worker.Kgworker{
			Matchfunc: matcher.Match,
			Keyfunc:   kg,
		}
		workers = append(workers, w)
	}
	wp := workerpool.WorkerPool[chan keygen.SSHKey]{
		Workers: workers,
		Results: make(chan keygen.SSHKey),
	}

	if c.StatsLogInterval != 0 {
		ticker := time.NewTicker(c.StatsLogInterval)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				printStats(wp.GetStats())
			}
		}()
	}

	wp.Start()
	result := <-wp.Results
	wps := wp.GetStats()
	printStats(wps)
	outputKey(c, wps.Elapsed, result)
}

func main() {
	ml := matcher.MatcherList()
	kl := keygen.KeygenList()

	commit := "none"
	date := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			if kv.Key == "vcs.revision" {
				commit = kv.Value
			} else if kv.Key == "vcs.time" {
				date = kv.Value
			}
		}
	}

	c := cli{}
	kongctx := kong.Parse(&c, kong.Vars{
		"version":         fmt.Sprintf("%s commit:%s date:%s goVersion:%s platform:%s/%s", version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		"default_threads": fmt.Sprintf("%d", runtime.NumCPU()),
		"keytypes":        strings.Join(kl.Names(), ","),
		"default_keytype": kl.Names()[0],
		"matchers":        strings.Join(ml.Names(), ","),
		"default_matcher": ml.Names()[0],
	})

	ctx := context.Background()
	if c.Metrics {
		exporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			slog.Error("Could not create metric exporter", "error", err)
			os.Exit(1)
		}
		resource, err := resource.Merge(resource.Default(),
			resource.NewWithAttributes(semconv.SchemaURL,
				semconv.ServiceName(serviceName),
				semconv.ServiceVersion(version),
				semconv.ServiceInstanceID(instanceID),
			))
		if err != nil {
			slog.Error("Could not crete metric resources", "error", err)
			os.Exit(1)
		}
		provider := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exporter)),
			metric.WithResource(resource),
		)

		otel.SetMeterProvider(provider)
	}

	if c.Profile {
		f, err := os.Create("./pprof")
		if err != nil {
			slog.Error("Could not create profile file", "error", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("Could not start profiling", "error", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
		// listening OS shutdown singal
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signalChan
			slog.Info("Got shutdown signal.")
			pprof.StopCPUProfile()
			os.Exit(1)
		}()
	}
	if c.PyroscopeProfile {
		profiler, err := pyroscope.Start(pyroscope.Config{
			Logger:          pyroscope.StandardLogger,
			ApplicationName: serviceName,
			// ServerAddress:     os.Getenv("PYROSCOPE_SERVER"),
			BasicAuthUser:     os.Getenv("GRAFANA_INSTANCE_ID"),
			BasicAuthPassword: os.Getenv("TOKEN"),
			Tags: map[string]string{
				"instanceID": instanceID,
				"platform":   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				"cpus":       fmt.Sprintf("%d", runtime.NumCPU()),
			},
		})
		if err != nil {
			slog.Error("Could not start profiling", "error", err)
			os.Exit(1)
		}
		defer func() {
			_ = profiler.Stop()
		}()
		// listening OS shutdown singal
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signalChan
			slog.Info("Got shutdown signal.")
			_ = profiler.Stop()
			os.Exit(1)
		}()
	}

	m, err := ml.Get(c.Matcher)
	if err != nil {
		slog.Error("Invalid matcher", "error", err)
		os.Exit(1)
	}
	m.SetMatchString(c.MatchString)

	k, ok := kl.Get(c.KeyType)
	if !ok {
		slog.Error("Invalid key type", "error", err)
		os.Exit(1)
	}
	runKeygen(c, m, k)
	kongctx.Exit(0)
}

func printStats(wps *workerpool.WorkerPoolStats) {
	slog.Info("Tested keys",
		slog.Duration("time", wps.Elapsed),
		slog.Int64("tested", wps.Count),
		slog.Float64("kKeys/s", float64(wps.Count)/wps.Elapsed.Seconds()/1000),
	)
}

func outputKey(c cli, elapsed time.Duration, result keygen.SSHKey) {
	privK := result.SSHPrivkey()
	pubK := result.SSHPubkey()
	slog.Info("Found matching public key", "pubkey", string(pubK))

	outDir := c.OutputDir + "/"
	if c.Output == "pem-files" {
		privkeyFileName := outDir + c.MatchString
		pubkeyFileName := outDir + c.MatchString + ".pub"
		_ = os.WriteFile(privkeyFileName, privK, 0o600)
		_ = os.WriteFile(pubkeyFileName, pubK, 0o600)
		slog.Info("Result keypair stored", "privkey_file", privkeyFileName, "pubkey_file", pubkeyFileName)
	} else if c.Output == "json-file" {
		file, _ := json.MarshalIndent(OutputData{
			PublicKey:  string(pubK),
			PrivateKey: string(privK),
			Metadata: Metadata{
				FindString: c.MatchString,
				Time:       int64(elapsed / time.Second),
			},
		}, "", " ")
		jsonFileName := outDir + "result.json"
		_ = os.WriteFile(jsonFileName, file, 0o600)
		slog.Info("Result keypair stored", "json_file", jsonFileName)
	}
}
