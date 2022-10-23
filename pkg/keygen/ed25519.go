package keygen

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/ed25519"
)

type ed25519Keygen struct {
}

func NewEd25519() *ed25519Keygen {
	return &ed25519Keygen{}
}

func (k *ed25519Keygen) Name() string {
	return "ed25519"
}

func (k *ed25519Keygen) New() sshkey.SSHKey {
	return ed25519.Init()
}
