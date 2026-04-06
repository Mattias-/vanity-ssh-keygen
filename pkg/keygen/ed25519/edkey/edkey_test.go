package edkey

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestMarshalED25519PrivateKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	marshaled := MarshalED25519PrivateKey(priv)
	if marshaled == nil {
		t.Fatal("MarshalED25519PrivateKey returned nil")
	}

	b := &pem.Block{
		Type:    "OPENSSH PRIVATE KEY",
		Headers: nil,
		Bytes:   marshaled,
	}
	privatePEM := pem.EncodeToMemory(b)

	_, err = ssh.ParseRawPrivateKey(privatePEM)
	if err != nil {
		t.Errorf("Failed to parse marshaled private key: %v", err)
	}
}
