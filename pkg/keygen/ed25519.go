package keygen

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

type ed25519Keygen struct {
}

func NewEd25519() *ed25519Keygen {
	return &ed25519Keygen{}
}

func (k *ed25519Keygen) Name() string {
	return "ed25519"
}

func (k *ed25519Keygen) New() key.SSHKey {
	return ed25519.Init()
}
