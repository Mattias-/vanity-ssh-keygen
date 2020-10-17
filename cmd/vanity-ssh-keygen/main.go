package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"time"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/worker"
)

func main() {
	threads := flag.Int("threads", runtime.NumCPU(), "threads")
	flag.Parse()
	var findString = flag.Args()[0]
	var outDir = "./"

	wp := worker.NewWorkerPool(
		*threads,
		matcher.LowercaseMatcher{MatchString: findString},
		ed25519.New,
	)
	wp.Start()

	// Log stats during execution
	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for range ticker.C {
			printStats(wp)
		}
	}()

	result := <-wp.Results

	ticker.Stop()
	printStats(wp)

	pubK, _ := (*result).SSHPubkey()
	privK, _ := (*result).SSHPrivkey()

	log.Println("Found pubkey:")
	log.Print(string(pubK))

	_ = ioutil.WriteFile(outDir+findString, privK, 0600)
	_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
}

func printStats(wp *worker.WorkerPool) {
	wps := wp.GetStats()
	log.Println("Time:", wps.Elapsed.Truncate(time.Second).String())
	log.Println("Tested:", wps.Count)
	log.Println(fmt.Sprintf("%.2f", float64(wps.Count)/wps.Elapsed.Seconds()/1000), "kKeys/s")
}
