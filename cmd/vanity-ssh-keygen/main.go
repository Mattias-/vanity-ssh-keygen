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
	"strconv"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

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

type config struct {
	Version          kong.VersionFlag `help:"Print version and exit"`
	Debug            bool             `help:"Enable debug logging" default:"false"`
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

type app struct {
	config        config
	shutdownFuncs []func(context.Context) error
}

type resultSink = func(elapsed time.Duration, result keygen.SSHKey)

func (a *app) shutdownAll() {
	slog.Debug("Shutting down", "funcs", len(a.shutdownFuncs))

	// Call shutdown functions with a new context with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var err error
	for _, fn := range a.shutdownFuncs {
		err = errors.Join(err, fn(shutdownCtx))
	}
	if err != nil {
		slog.Error("Error shutting down", "error", err)
		// Not critical so still exit with code 0.
	}
}

func (a *app) addShutdownFunc(fn func(context.Context) error) {
	a.shutdownFuncs = append(a.shutdownFuncs, fn)
}

func main() {
	matcher.RegisterMatcher("ignorecase", ignorecase.New())
	matcher.RegisterMatcher("ignorecase-ed25519", ignorecaseed25519.New())
	keygen.RegisterKeygen("ed25519", func() keygen.SSHKey { return ed25519.New() })
	keygen.RegisterKeygen("rsa-2048", func() keygen.SSHKey { return rsa.New(2048) })
	keygen.RegisterKeygen("rsa-4096", func() keygen.SSHKey { return rsa.New(4096) })

	defaultThreads := runtime.NumCPU()
	overrideThreads := os.Getenv("OVERRIDE_DEFAULT_THREADS")
	if overrideThreads != "" {
		defaultThreads, _ = strconv.Atoi(overrideThreads)
	}
	var a app
	_ = kong.Parse(&a.config, kong.Vars{
		"version":         versionString(),
		"default_threads": fmt.Sprintf("%d", defaultThreads),
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

	a.addShutdownFunc(func(ctx context.Context) error {
		slog.Debug("Stopping signal handler")
		stop()
		return nil
	})

	if a.config.Metrics {
		exporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			slog.Error("Could not create metric exporter", "error", err)
			os.Exit(1)
		}
		res := resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.ServiceInstanceID(instanceID),
		)
		provider := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exporter)),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(provider)

		a.addShutdownFunc(func(ctx context.Context) error {
			slog.Debug("Shutting down metric provider")
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to stop metric provider: %w", err)
			}
			return nil
		})
	}

	logLevel := slog.LevelInfo
	if a.config.Debug {
		logLevel = slog.LevelDebug
	}
	if a.config.OtelLogs {
		exporter, err := otlploghttp.New(ctx)
		if err != nil {
			slog.Error("Could not create log exporter", "error", err)
			os.Exit(1)
		}
		provider := log.NewLoggerProvider(
			log.WithProcessor(
				log.NewBatchProcessor(exporter),
			))
		global.SetLoggerProvider(provider)
		slog.SetDefault(otelslog.NewLogger(serviceName,
			otelslog.WithAttributes(
				semconv.ServiceVersion(version),
				semconv.ServiceInstanceID(instanceID),
			),
		))

		a.addShutdownFunc(func(ctx context.Context) error {
			slog.Debug("Shutting down logger provider")
			if err := provider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to stop logger provider: %w", err)
			}
			return nil
		})
	} else {
		slog.SetDefault(
			slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: logLevel,
			})),
		)
	}

	if a.config.Profile {
		f, err := os.Create("./pprof")
		if err != nil {
			slog.Error("Could not create profile file", "error", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("Could not start profiling", "error", err)
			os.Exit(1)
		}

		a.addShutdownFunc(func(_ context.Context) error {
			slog.Debug("Stopping CPU profile")
			pprof.StopCPUProfile()
			return nil
		})
	}

	if a.config.PyroscopeProfile {
		profiler, err := pyroscope.Start(pyroscope.Config{
			Logger: func() pyroscope.Logger {
				if a.config.Debug {
					return pyroscope.StandardLogger
				}
				return nil
			}(),
			ApplicationName: serviceName,
			// ServerAddress:     os.Getenv("PYROSCOPE_SERVER"),
			BasicAuthUser:     os.Getenv("GRAFANA_INSTANCE_ID"),
			BasicAuthPassword: os.Getenv("TOKEN"),
			Tags: map[string]string{
				"instanceID":         instanceID,
				"platform":           fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				"cpus":               fmt.Sprintf("%d", runtime.NumCPU()),
				"service_git_ref":    fmt.Sprintf("v%s", version),
				"service_repository": "https://github.com/Mattias-/vanity-ssh-keygen",
			},
		})
		if err != nil {
			slog.Error("Could not start profiling", "error", err)
			os.Exit(1)
		}

		a.addShutdownFunc(func(_ context.Context) error {
			slog.Debug("Stopping profiler")
			return profiler.Stop()
		})
	}

	m, ok := matcher.Get(a.config.Matcher)
	if !ok {
		slog.Error("Invalid matcher")
		os.Exit(1)
	}
	m.SetMatchString(a.config.MatchString)
	k, ok := keygen.Get(a.config.KeyType)
	if !ok {
		slog.Error("Invalid key type")
		os.Exit(1)
	}

	var outputter resultSink
	switch a.config.Output {
	case "pem-files":
		outputter = a.outputPEM
	case "json-file":
		outputter = a.outputJSON
	default:
		slog.Error("Invalid output format", "output", a.config.Output)
	}

	go func() {
		<-ctx.Done()
		a.shutdownAll()
		os.Exit(1)
	}()
	a.runKeygen(m, k, outputter)
	a.shutdownAll()
	os.Exit(0)
}

func (a *app) runKeygen(matcher matcher.Matcher, kg keygen.Keygen, outputter resultSink) {
	wp := workerpool.WorkerPool[chan keygen.SSHKey]{
		Workers: make([]workerpool.Worker[chan keygen.SSHKey], 0, a.config.Threads),
		Results: make(chan keygen.SSHKey),
	}
	for range a.config.Threads {
		wp.Workers = append(wp.Workers, &keygen.Worker{
			Matchfunc: matcher.Match,
			Keyfunc:   kg,
		})
	}

	if a.config.StatsLogInterval != 0 {
		ticker := time.NewTicker(a.config.StatsLogInterval)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				wp.GetStats().Log()
			}
		}()
	}

	wp.Start()
	result := <-wp.Results
	wps := wp.GetStats()
	wps.Log()

	outputter(wps.Elapsed, result)
}

func (a *app) outputPEM(elapsed time.Duration, result keygen.SSHKey) {
	privK := result.SSHPrivkey()
	pubK := result.SSHPubkey()
	slog.Info("Found matching public key", "pubkey", string(pubK))
	outDir := a.config.OutputDir + "/"

	privkeyFileName := outDir + a.config.MatchString
	pubkeyFileName := outDir + a.config.MatchString + ".pub"
	_ = os.WriteFile(privkeyFileName, privK, 0o600)
	_ = os.WriteFile(pubkeyFileName, pubK, 0o600)
	slog.Info("Result keypair stored",
		"privkey_file", privkeyFileName,
		"pubkey_file", pubkeyFileName,
		slog.Duration("elapsed", elapsed),
	)
}

func (a *app) outputJSON(elapsed time.Duration, result keygen.SSHKey) {
	privK := result.SSHPrivkey()
	pubK := result.SSHPubkey()
	slog.Info("Found matching public key", "pubkey", string(pubK))
	outDir := a.config.OutputDir + "/"

	file, _ := json.MarshalIndent(OutputData{
		PublicKey:  string(pubK),
		PrivateKey: string(privK),
		Metadata: Metadata{
			FindString: a.config.MatchString,
			Time:       int64(elapsed / time.Second),
		},
	}, "", " ")
	jsonFileName := outDir + "result.json"
	_ = os.WriteFile(jsonFileName, file, 0o600)
	slog.Info("Result keypair stored", "json_file", jsonFileName)
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
