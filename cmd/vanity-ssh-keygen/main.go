package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

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
		fmt.Println(viper.AllSettings())
		var findString = args[0]
		var outDir = "./"

		var wp *worker.WorkerPool

		if viper.GetBool("metricServer") {
			go func() {
				http.Handle("/metrics", promhttp.Handler())
				log.Fatal(http.ListenAndServe(viper.GetString("metricsListen"), nil))
			}()
		}

		if viper.GetBool("logStats") {
			ticker := time.NewTicker(time.Second * 2)
			defer ticker.Stop()
			go func() {
				for range ticker.C {
					printStats(wp)
				}
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

		privK, _ := (*result).SSHPrivkey()
		_ = ioutil.WriteFile(outDir+findString, privK, 0600)
		pubK, _ := (*result).SSHPubkey()
		_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
		log.Print("Found pubkey: ", string(pubK))

	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
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

	viper.SetDefault("logStats", true)
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
