package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/mikesmitty/edkey"
)

var workers []*Worker

func main() {
	threads := flag.Int("threads", runtime.NumCPU(), "threads")
	flag.Parse()
	var findString = flag.Args()[0]
	var outDir = "./"

	results := make(chan SSHKey)
	lm := LowercaseMatcher{findString}
	for i := 0; i < *threads; i++ {
		w := &Worker{
			Count:        0,
			Matcher:      lm,
			KeyGenerator: NewED,
		}
		workers = append(workers, w)
		go w.findKey(results)
	}

	// Log stats during execution
	start := time.Now()
	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for range ticker.C {
			printStats(start)
		}
	}()

	result := <-results

	ticker.Stop()
	printStats(start)

	pubK, _ := result.SSHPubkey()
	privK, _ := result.SSHPrivkey()

	log.Println("Found pubkey:")
	log.Print(string(pubK))

	_ = ioutil.WriteFile(outDir+findString, privK, 0600)
	_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
}

func printStats(start time.Time) {
	sum := totalCount()
	elapsed := time.Since(start)
	log.Println("Time:", elapsed.Truncate(time.Second).String())
	log.Println("Tested:", sum)
	log.Println(fmt.Sprintf("%.2f", sum/elapsed.Seconds()/1000), "kKeys/s")
}

func totalCount() float64 {
	var sum float64
	for _, w := range workers {
		sum += float64(w.Count)
	}
	return sum
}

type Matcher interface {
	Match(SSHKey) bool
}

type LowercaseMatcher struct {
	matchString string
}

func (m LowercaseMatcher) Match(s SSHKey) bool {
	if s == nil {
		return false
	}
	pubK, err := s.SSHPubkey()
	if err != nil {
		log.Println(err)
		return false
	}
	return strings.Contains(strings.ToLower(string(pubK)), m.matchString)
}

type Worker struct {
	Count uint64
	Matcher
	KeyGenerator func() SSHKey
}

func (w *Worker) findKey(result chan SSHKey) {
	var err error
	var k SSHKey
	for {
		w.Count += 1
		k = w.KeyGenerator()
		if err != nil {
			log.Println(err)
			return
		}
		if w.Matcher.Match(k) {
			break
		}
	}
	result <- k
}

type SSHKey interface {
	SSHPubkey() ([]byte, error)
	SSHPrivkey() ([]byte, error)
}

type ed struct {
	publicKey  crypto.PublicKey
	privateKey ed25519.PrivateKey
}

func NewED() SSHKey {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Println(err)
		return nil
	}
	return ed{publicKey, privateKey}
}

func (s ed) SSHPubkey() ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(s.publicKey)
	if err != nil {
		return nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}

func (s ed) SSHPrivkey() ([]byte, error) {
	privDER := edkey.MarshalED25519PrivateKey(s.privateKey)
	b := pem.Block{
		Type:    "OPENSSH PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&b)

	return privatePEM, nil
}
