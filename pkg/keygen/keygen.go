package keygen

import "github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"

type Keygen interface {
	Name() string
	New() key.SSHKey
}
