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

func New() key.SSHKey {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
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
