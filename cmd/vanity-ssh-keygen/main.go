package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/worker"
)

var configFile string

var matchers map[string]matcher.Matcher

var rootCmd = &cobra.Command{
	Use:   "vanity-ssh-keygen [match-string]",
	Short: "Generate a vanity SSH key that matches specified condition",
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

		wp = worker.NewWorkerPool(
			viper.GetInt("threads"),
			matcher.LowercaseMatcher{MatchString: findString},
			ed25519.New,
		)
		wp.Start()
		result := <-wp.Results

		printStats(wp)

		pubK, _ := (*result).SSHPubkey()
		privK, _ := (*result).SSHPrivkey()

		log.Println("Found pubkey:")
		log.Print(string(pubK))

		_ = ioutil.WriteFile(outDir+findString, privK, 0600)
		_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	//matchers[matcher.LowercaseMatcher.Name()] = matcher.LowercaseMatcher

	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(&configFile, "config", "", "Config file")

	rootCmd.Flags().IntP("threads", "j", runtime.NumCPU(), "Execution threads. Defaults to the number of logical CPU cores")
	_ = viper.BindPFlag("threads", rootCmd.Flags().Lookup("threads"))

	rootCmd.Flags().String("matcher", "lowercase", "Matcher used to find a vanity SSH key")
	_ = viper.BindPFlag("matcher", rootCmd.Flags().Lookup("matcher"))

	rootCmd.Flags().StringP("key-type", "t", "ed25519", "Key type to generate")
	_ = viper.BindPFlag("keyType", rootCmd.Flags().Lookup("key-type"))

	viper.SetDefault("logStats", true)
	viper.SetDefault("metricServer", false)
	viper.SetDefault("metricsListen", ":9090")

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
