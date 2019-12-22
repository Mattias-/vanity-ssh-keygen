package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	mathrand "math/rand"
	"runtime"
	"strings"
	"time"
)

var workers []*Worker

func main() {
	flag.Parse()
	var findString = flag.Args()[0]
	var outDir = "./"

	lm := LowercaseMatcher{findString}
	results := make(chan SSHKey)

	for i := 0; i < runtime.NumCPU(); i++ {
		w := &Worker{0}
		workers = append(workers, w)
		go w.findKey(NewED, lm, results)
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

	fmt.Print(string(pubK))
	//fmt.Print(string(privK))

	_ = ioutil.WriteFile(outDir+findString, privK, 0600)
	_ = ioutil.WriteFile(outDir+findString+".pub", pubK, 0644)
}

func printStats(start time.Time) {
	sum := totalCount()
	elapsed := time.Since(start)
	fmt.Println("Time:", elapsed.String())
	fmt.Println("Tested:", sum)
	fmt.Println(fmt.Sprintf("%.2f", sum/elapsed.Seconds()/1000), "kKeys/s")
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
		fmt.Println(err)
		return false
	}
	return strings.Contains(strings.ToLower(string(pubK)), m.matchString)
}

type Worker struct {
	Count uint64
}

func (w *Worker) findKey(generator func() SSHKey, matcher Matcher, result chan SSHKey) {
	var err error
	var k SSHKey
	for {
		w.Count += 1
		k = generator()
		if err != nil {
			fmt.Println(err)
			return
		}
		if matcher.Match(k) {
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
		fmt.Println(err)
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
	privDER := MarshalED25519PrivateKey(s.privateKey)
	b := pem.Block{
		Type:    "OPENSSH PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&b)

	return privatePEM, nil
}

/* Writes ed25519 private keys into the new OpenSSH private key format.
I have no idea why this isn't implemented anywhere yet, you can do seemingly
everything except write it to disk in the OpenSSH private key format. */
func MarshalED25519PrivateKey(key ed25519.PrivateKey) []byte {
	// Add our key header (followed by a null byte)
	magic := append([]byte("openssh-key-v1"), 0)

	var w struct {
		CipherName   string
		KdfName      string
		KdfOpts      string
		NumKeys      uint32
		PubKey       []byte
		PrivKeyBlock []byte
	}

	// Fill out the private key fields
	pk1 := struct {
		Check1  uint32
		Check2  uint32
		Keytype string
		Pub     []byte
		Priv    []byte
		Comment string
		Pad     []byte `ssh:"rest"`
	}{}

	// Set our check ints
	ci := mathrand.Uint32()
	pk1.Check1 = ci
	pk1.Check2 = ci

	// Set our key type
	pk1.Keytype = ssh.KeyAlgoED25519

	// Add the pubkey to the optionally-encrypted block
	pk, ok := key.Public().(ed25519.PublicKey)
	if !ok {
		fmt.Println("ed25519.PublicKey type assertion failed on an ed25519 public key. This should never ever happen.")
		return nil
	}
	pubKey := []byte(pk)
	pk1.Pub = pubKey

	// Add our private key
	pk1.Priv = []byte(key)

	// Might be useful to put something in here at some point
	pk1.Comment = ""

	// Add some padding to match the encryption block size within PrivKeyBlock (without Pad field)
	// 8 doesn't match the documentation, but that's what ssh-keygen uses for unencrypted keys. *shrug*
	bs := 8
	blockLen := len(ssh.Marshal(pk1))
	padLen := (bs - (blockLen % bs)) % bs
	pk1.Pad = make([]byte, padLen)

	// Padding is a sequence of bytes like: 1, 2, 3...
	for i := 0; i < padLen; i++ {
		pk1.Pad[i] = byte(i + 1)
	}

	// Generate the pubkey prefix "\0\0\0\nssh-ed25519\0\0\0 "
	prefix := []byte{0x0, 0x0, 0x0, 0x0b}
	prefix = append(prefix, []byte(ssh.KeyAlgoED25519)...)
	prefix = append(prefix, []byte{0x0, 0x0, 0x0, 0x20}...)

	// Only going to support unencrypted keys for now
	w.CipherName = "none"
	w.KdfName = "none"
	w.KdfOpts = ""
	w.NumKeys = 1
	w.PubKey = append(prefix, pubKey...)
	w.PrivKeyBlock = ssh.Marshal(pk1)

	magic = append(magic, ssh.Marshal(w)...)

	return magic
}
