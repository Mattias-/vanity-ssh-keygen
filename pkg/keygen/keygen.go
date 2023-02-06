package keygen

import (
	"golang.org/x/exp/maps"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/rsa"
)

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	New()
}

type Keygen func() SSHKey

type kli map[string]Keygen

func KeygenList() kli {
	return kli{
		"ed25519": func() SSHKey {
			return ed25519.Init()
		},
		"rsa-2048": func() SSHKey {
			return rsa.Init(2048)
		},
		"rsa-4096": func() SSHKey {
			return rsa.Init(4096)
		},
	}
}

func (kl kli) Names() []string {
	return maps.Keys(kl)
}

func (kl kli) Get(name string) (Keygen, bool) {
	v, ok := kl[name]
	return v, ok
}
