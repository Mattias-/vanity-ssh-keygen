package ed25519

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
	"github.com/mikesmitty/edkey"
)

type ed struct {
	publicKey  crypto.PublicKey
	privateKey ed25519.PrivateKey
}

func Init() key.SSHKey {
	return &ed{}
}

func (s *ed) New() {
	s.publicKey, s.privateKey, _ = ed25519.GenerateKey(rand.Reader)
}

func (s *ed) SSHPubkey() []byte {
	publicKey, _ := ssh.NewPublicKey(s.publicKey)
	return ssh.MarshalAuthorizedKey(publicKey)
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
