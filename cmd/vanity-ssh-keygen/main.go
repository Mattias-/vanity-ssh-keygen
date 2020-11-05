package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/worker"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	configFile string
	matchers   = make(map[string]matcher.Matcher)
	keygens    = make(map[string]keygen.Keygen)
)

var rootCmd = &cobra.Command{
	Use:   "vanity-ssh-keygen [match-string]",
	Short: "Generate a vanity SSH key that matches the input argument",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runKeygen(args[0])
	},
}

type Metadata struct {
	FindString string `json:"findstring"`
	Time       int64  `json:"time"`
}

type OutputData struct {
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_key"`
	Metadata   Metadata `json:"metadata"`
}

func runKeygen(findString string) {
	fmt.Println(viper.AllSettings())

	var wp *worker.WorkerPool

	if viper.GetBool("metricServer") {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Fatal(http.ListenAndServe(viper.GetString("metricsListen"), nil))
		}()
	}

	logStatsInterval := viper.GetInt64("logStatsInterval")
	if logStatsInterval != 0 {
		ticker := time.NewTicker(time.Second * time.Duration(logStatsInterval))
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				printStats(wp)
			}
		}()
	}

	if viper.GetBool("profile") {
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

	m := matchers[viper.GetString("matcher")]
	m.SetMatchString(findString)

	kg := keygens[viper.GetString("keyType")]

	wp = worker.NewWorkerPool(
		viper.GetInt("threads"),
		m,
		kg,
	)
	wp.Start()
	result := <-wp.Results

	printStats(wp)

	privK := (*result).SSHPrivkey()
	pubK := (*result).SSHPubkey()
	log.Print("Found pubkey: ", string(pubK))

	outDir := viper.GetString("outputDirectory")
	output := viper.GetString("output")
	if output == "pem-files" {
		_ = ioutil.WriteFile(outDir+findString, privK, 0600)
		_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
		log.Printf("Keypair written to: %[1]s and %[1]s.pub", outDir+findString)
	} else if output == "json-file" {
		file, _ := json.MarshalIndent(OutputData{
			PublicKey:  string(pubK),
			PrivateKey: string(privK),
			Metadata: Metadata{
				FindString: findString,
				Time:       int64(wp.GetStats().Elapsed / time.Second),
			},
		}, "", " ")
		_ = ioutil.WriteFile(outDir+"result.json", file, 0600)
		log.Printf("Result written to: %s", outDir+"result.json")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		pprof.StopCPUProfile()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	lm := matcher.NewIgnorecaseMatcher()
	matchers[lm.Name()] = lm

	lmEd25519 := matcher.NewIgnorecaseEd25519Matcher()
	matchers[lmEd25519.Name()] = lmEd25519

	ed25519 := keygen.NewEd25519()
	keygens[ed25519.Name()] = ed25519

	rsa2048 := keygen.NewRsa(2048)
	keygens[rsa2048.Name()] = rsa2048

	rsa4096 := keygen.NewRsa(4096)
	keygens[rsa4096.Name()] = rsa4096

	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(&configFile, "config", "", "Config file")

	rootCmd.Flags().IntP("threads", "j", runtime.NumCPU(), "Execution threads. Defaults to the number of logical CPU cores")
	_ = viper.BindPFlag("threads", rootCmd.Flags().Lookup("threads"))

	rootCmd.Flags().String("matcher", "ignorecase", "Matcher used to find a vanity SSH key")
	_ = viper.BindPFlag("matcher", rootCmd.Flags().Lookup("matcher"))

	rootCmd.Flags().StringP("key-type", "t", "ed25519", "Key type to generate")
	_ = viper.BindPFlag("keyType", rootCmd.Flags().Lookup("key-type"))

	rootCmd.Flags().Bool("profile", false, "Write pprof CPU profile to ./pprof")
	_ = viper.BindPFlag("profile", rootCmd.Flags().Lookup("profile"))

	rootCmd.Flags().StringP("output", "o", "pem-files", "Output format.")
	_ = viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))

	rootCmd.Flags().String("output-dir", "./", "Output directory.")
	_ = viper.BindPFlag("outputDirectory", rootCmd.Flags().Lookup("output-dir"))

	viper.SetDefault("logStatsInterval", 2)
	viper.SetDefault("metricServer", false)
	viper.SetDefault("metricsListen", ":9101")

	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s %s %s", version, commit, date))
}

func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	}

	viper.SetEnvPrefix("vskg")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func printStats(wp *worker.WorkerPool) {
	wps := wp.GetStats()
	log.Println("Time:", wps.Elapsed.Truncate(time.Second).String())
	log.Println("Tested:", wps.Count)
	log.Println(fmt.Sprintf("%.2f", float64(wps.Count)/wps.Elapsed.Seconds()/1000), "kKeys/s")
}
