package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/worker"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/workerpool"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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
	Metrics          bool             `help:"Enable metrics server." default:"true" negatable:""`
	MetricsPort      int              `help:"Listening port for metrics server." default:"9101"`
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
	ctx := kong.Parse(&c, kong.Vars{
		"version":         fmt.Sprintf("%s commit:%s date:%s goVersion:%s platform:%s/%s", version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		"default_threads": fmt.Sprintf("%d", runtime.NumCPU()),
		"keytypes":        strings.Join(kl.Names(), ","),
		"default_keytype": kl.Names()[0],
		"matchers":        strings.Join(ml.Names(), ","),
		"default_matcher": ml.Names()[0],
	})

	if c.Metrics {
		exporter, err := prometheus.New()
		if err != nil {
			log.Fatal(err)
		}
		provider := metric.NewMeterProvider(
			metric.WithReader(exporter),
			metric.WithResource(
				resource.NewWithAttributes(semconv.SchemaURL,
					semconv.ServiceNameKey.String("vanity-ssh-keygen"),
					semconv.ServiceVersionKey.String(version),
					semconv.ServiceInstanceIDKey.String(uuid.NewString()),
				)))

		global.SetMeterProvider(provider)
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.MetricsPort), nil))
		}()
	}

	if c.Profile {
		f, err := os.Create("./pprof")
		if err != nil {
			log.Fatal("Could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("Could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
		// listening OS shutdown singal
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signalChan
			log.Printf("Got shutdown signal.")
			pprof.StopCPUProfile()
			os.Exit(1)
		}()
	}

	m, err := ml.Get(c.Matcher)
	if err != nil {
		log.Fatal("Invalid matcher")
	}
	m.SetMatchString(c.MatchString)

	k, ok := kl.Get(c.KeyType)
	if !ok {
		log.Fatal("Invalid key type")
	}
	runKeygen(c, m, k)
	ctx.Exit(0)
}

func printStats(wps *workerpool.WorkerPoolStats) {
	log.Println("Time:", wps.Elapsed.Truncate(time.Second).String())
	log.Println("Tested:", wps.Count)
	log.Println(fmt.Sprintf("%.2f", float64(wps.Count)/wps.Elapsed.Seconds()/1000), "kKeys/s")
}

func outputKey(c cli, elapsed time.Duration, result keygen.SSHKey) {
	privK := result.SSHPrivkey()
	pubK := result.SSHPubkey()
	log.Print("Found pubkey: ", string(pubK))

	outDir := c.OutputDir + "/"
	if c.Output == "pem-files" {
		_ = os.WriteFile(outDir+c.MatchString, privK, 0600)
		_ = os.WriteFile(outDir+c.MatchString+".pub", pubK, 0644)
		log.Printf("Keypair written to: %[1]s and %[1]s.pub", outDir+c.MatchString)
	} else if c.Output == "json-file" {
		file, _ := json.MarshalIndent(OutputData{
			PublicKey:  string(pubK),
			PrivateKey: string(privK),
			Metadata: Metadata{
				FindString: c.MatchString,
				Time:       int64(elapsed / time.Second),
			},
		}, "", " ")
		_ = os.WriteFile(outDir+"result.json", file, 0600)
		log.Printf("Result written to: %s", outDir+"result.json")
	}
}
