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
	publicKey  rsa.PublicKey
	privateKey rsa.PrivateKey
}

func New(bitSize int) key.SSHKey {
	privateKey, _ := rsa.GenerateKey(rand.Reader, bitSize)
	return localRsa{privateKey.PublicKey, *privateKey}
}

func (s localRsa) SSHPubkey() []byte {
	publicKey, _ := ssh.NewPublicKey(&s.publicKey)
	return ssh.MarshalAuthorizedKey(publicKey)
}

func (s localRsa) SSHPrivkey() []byte {
	privDER := x509.MarshalPKCS1PrivateKey(&s.privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}
