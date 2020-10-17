package matcher

import "github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"

type Matcher interface {
	Match(*key.SSHKey) bool
	Name() string
}
