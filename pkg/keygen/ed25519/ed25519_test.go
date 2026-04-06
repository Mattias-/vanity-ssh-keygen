package ed25519

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestEd25519(t *testing.T) {
	e := New()
	e.Generate()

	pub := e.SSHPubkey()
	if len(pub) == 0 {
		t.Error("SSHPubkey() returned empty result")
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey(pub)
	if err != nil {
		t.Errorf("Failed to parse authorized key: %v", err)
	}

	priv := e.SSHPrivkey()
	if len(priv) == 0 {
		t.Error("SSHPrivkey() returned empty result")
	}

	_, err = ssh.ParseRawPrivateKey(priv)
	if err != nil {
		t.Errorf("Failed to parse private key: %v", err)
	}
}
