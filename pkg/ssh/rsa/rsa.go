package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

type localRsa struct {
	privateKey *rsa.PrivateKey
	bitSize    int
}

func Init(bits int) key.SSHKey {
	return &localRsa{bitSize: bits}
}

func (s *localRsa) New() {
	s.privateKey, _ = rsa.GenerateKey(rand.Reader, s.bitSize)
}

func (s *localRsa) SSHPubkey() []byte {
	publicKey, _ := ssh.NewPublicKey(&s.privateKey.PublicKey)
	return ssh.MarshalAuthorizedKey(publicKey)
}

func (s *localRsa) SSHPrivkey() []byte {
	privDER := x509.MarshalPKCS1PrivateKey(s.privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}
