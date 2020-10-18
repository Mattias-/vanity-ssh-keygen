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
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}
	return localRsa{privateKey.PublicKey, *privateKey}
}

func (s localRsa) SSHPubkey() ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(&s.publicKey)
	if err != nil {
		return nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}

func (s localRsa) SSHPrivkey() ([]byte, error) {
	privDER := x509.MarshalPKCS1PrivateKey(&s.privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM, nil
}
