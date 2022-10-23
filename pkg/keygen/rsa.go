package keygen

import (
	"fmt"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey/rsa"
)

type rsaKeygen struct {
	name string
	bits int
}

func NewRsa(bits int) *rsaKeygen {
	return &rsaKeygen{
		name: "rsa-" + fmt.Sprintf("%d", bits),
		bits: bits,
	}
}

func (k *rsaKeygen) Name() string {
	return k.name
}

func (k *rsaKeygen) New() sshkey.SSHKey {
	return rsa.Init(k.bits)
}
