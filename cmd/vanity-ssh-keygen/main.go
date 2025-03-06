package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen/rsa"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher/ignorecase"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher/ignorecaseed25519"
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
	OtelLogs         bool             `help:"Enable otel logs." default:"false"`
	Output           string           `short:"o" help:"Output format. One of: pem-files|json-file." default:"pem-files"`
	OutputDir        string           `help:"Output directory." default:"./" type:"existingdir"`
	StatsLogInterval time.Duration    `help:"Statistics will be printed at this interval, set to 0 to disable" default:"2s"`
}

func main() {
	matcher.RegisterMatcher("ignorecase", ignorecase.New())
	matcher.RegisterMatcher("ignorecase-ed25519", ignorecaseed25519.New())
	keygen.RegisterKeygen("ed25519", func() keygen.SSHKey { return ed25519.New() })
	keygen.RegisterKeygen("rsa-2048", func() keygen.SSHKey { return rsa.New(2048) })
	keygen.RegisterKeygen("rsa-4096", func() keygen.SSHKey { return rsa.New(4096) })

	c := cli{}
	kongctx := kong.Parse(&c, kong.Vars{
		"version":         versionString(),
		"default_threads": fmt.Sprintf("%d", runtime.NumCPU()),
		"keytypes":        strings.Join(keygen.Names(), ","),
		"default_keytype": keygen.Names()[0],
		"matchers":        strings.Join(matcher.Names(), ","),
		"default_matcher": matcher.Names()[0],
	})

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
	)

	shutdownFuncs := []func(context.Context) error{}

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

		shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
			slog.Info("Shutting down metric provider")
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to stop metric provider: %w", err)
			}
			return nil
		})

		otel.SetMeterProvider(provider)
	}

	if c.OtelLogs {
		exporter, err := otlploghttp.New(ctx)
		if err != nil {
			panic(err)
		}

		provider := log.NewLoggerProvider(
			log.WithProcessor(
				log.NewBatchProcessor(exporter),
			))

		shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
			slog.Info("Shutting down logger provider")
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to stop logger provider: %w", err)
			}
			return nil
		})

		global.SetLoggerProvider(provider)

		slog.SetDefault(otelslog.NewLogger(serviceName,
			otelslog.WithSchemaURL(semconv.SchemaURL),
			otelslog.WithVersion(version),
		))

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

		shutdownFuncs = append(shutdownFuncs, func(_ context.Context) error {
			slog.Info("Stopping CPU profile")
			pprof.StopCPUProfile()
			return nil
		})

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

		shutdownFuncs = append(shutdownFuncs, func(_ context.Context) error {
			slog.Info("Stopping profiler")
			return profiler.Stop()
		})
	}

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down", "funcs", len(shutdownFuncs))

		// Call shutdown functions with a new context with timeout.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(shutdownCtx))
		}
		fmt.Printf("err: %v\n", err)
		kongctx.Exit(1)
	}()

	m, ok := matcher.Get(c.Matcher)
	if !ok {
		slog.Error("Invalid matcher")
		os.Exit(1)
	}
	m.SetMatchString(c.MatchString)
	k, ok := keygen.Get(c.KeyType)
	if !ok {
		slog.Error("Invalid key type")
		os.Exit(1)
	}
	runKeygen(c, m, k)

	stop()
	kongctx.Exit(0)
}

func runKeygen(c cli, matcher matcher.Matcher, kg keygen.Keygen) {
	var workers []workerpool.Worker[chan keygen.SSHKey]
	for i := 0; i < c.Threads; i++ {
		w := &keygen.Worker{
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

func versionString() string {
	commit := "none"
	date := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				commit = kv.Value
			case "vcs.time":
				date = kv.Value
			}
		}
	}

	return fmt.Sprintf("%s commit:%s date:%s goVersion:%s platform:%s/%s",
		version,
		commit,
		date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
