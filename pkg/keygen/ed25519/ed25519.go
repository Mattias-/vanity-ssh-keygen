package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen/ed25519/edkey"
)

var ed25519BinaryHeader = []byte{0, 0, 0, 11, 's', 's', 'h', '-', 'e', 'd', '2', '5', '5', '1', '9', 0, 0, 0, 32}

type ed struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	pubKeyBuf  [81]byte
}

func New() *ed {
	return &ed{}
}

func (s *ed) Generate() {
	s.publicKey, s.privateKey, _ = ed25519.GenerateKey(rand.Reader)
	s.updatePubkey()
}

func (s *ed) updatePubkey() {
	var bin [51]byte
	copy(bin[0:19], ed25519BinaryHeader)
	copy(bin[19:51], s.publicKey)

	copy(s.pubKeyBuf[0:12], "ssh-ed25519 ")
	base64.StdEncoding.Encode(s.pubKeyBuf[12:80], bin[:])
	s.pubKeyBuf[80] = '\n'
}

func (s *ed) SSHPubkey() []byte {
	return s.pubKeyBuf[:]
}

func (s *ed) SSHPrivkey() []byte {
	privDER := edkey.MarshalED25519PrivateKey(s.privateKey)
	b := pem.Block{
		Type:    "OPENSSH PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&b)
	return privatePEM
}
