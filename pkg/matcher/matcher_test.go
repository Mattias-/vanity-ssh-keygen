package matcher

import (
	"math/rand/v2"
	"testing"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher/ignorecase"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher/ignorecaseed25519"
)

type testKey []byte

func (k testKey) SSHPubkey() []byte {
	return k
}

func (k testKey) SSHPrivkey() []byte {
	panic("not implemented") // TODO: Implement
}

func (k testKey) Generate() {
	panic("not implemented") // TODO: Implement
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		// #nosec G404
		b[i] = letters[rand.IntN(len(letters))]
	}
	return b
}

func BenchmarkMatchIgnorecase(b *testing.B) {
	m := ignorecase.New()
	m.SetMatchString("abcdef")
	var k testKey
	b.ResetTimer()

	for b.Loop() {
		k = randSeq(100)
		m.Match(k)
	}
}

func BenchmarkMatchIgnorecaseED25519(b *testing.B) {
	m := ignorecaseed25519.New()
	m.SetMatchString("abcdef")
	var k testKey
	b.ResetTimer()

	for b.Loop() {
		k = randSeq(100)
		m.Match(k)
	}
}
