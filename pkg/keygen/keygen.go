package keygen

import (
	"errors"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
)

type Keygen interface {
	Name() string
	New() sshkey.SSHKey
}

func KeygenList() kli {
	return []Keygen{
		NewEd25519(),
		NewRsa(2048),
		NewRsa(4096),
	}
}

type kli []Keygen

func (kl kli) Names() []string {
	var ks []string
	for _, k := range kl {
		ks = append(ks, k.Name())
	}
	return ks
}

func (kl kli) Get(name string) (Keygen, error) {
	for _, k := range kl {
		if name == k.Name() {
			return k, nil
		}
	}
	return nil, errors.New("Unknown item")
}
