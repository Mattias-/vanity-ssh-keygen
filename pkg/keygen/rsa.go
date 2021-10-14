package keygen

import (
	"fmt"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/rsa"
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

func (k *rsaKeygen) New() key.SSHKey {
	return rsa.Init(k.bits)
}
