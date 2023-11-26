package keygen

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/rsa"
)

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	New()
}

type Keygen func() SSHKey

type kli []struct {
	name string
	f    Keygen
}

func KeygenList() kli {
	return kli{
		{"ed25519", func() SSHKey { return ed25519.Init() }},
		{"rsa-2048", func() SSHKey { return rsa.Init(2048) }},
		{"rsa-4096", func() SSHKey { return rsa.Init(4096) }},
	}
}

func (kl kli) Names() []string {
	r := make([]string, 0, len(kl))
	for _, k := range kl {
		r = append(r, k.name)
	}
	return r
}

func (kl kli) Get(name string) (Keygen, bool) {
	for _, k := range kl {
		if name == k.name {
			return k.f, true
		}
	}
	return nil, false
}
