package rsa

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestRSA(t *testing.T) {
	r := New(2048)
	r.Generate()

	pub := r.SSHPubkey()
	if len(pub) == 0 {
		t.Error("SSHPubkey() returned empty result")
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey(pub)
	if err != nil {
		t.Errorf("Failed to parse authorized key: %v", err)
	}

	priv := r.SSHPrivkey()
	if len(priv) == 0 {
		t.Error("SSHPrivkey() returned empty result")
	}

	_, err = ssh.ParseRawPrivateKey(priv)
	if err != nil {
		t.Errorf("Failed to parse private key: %v", err)
	}
}
