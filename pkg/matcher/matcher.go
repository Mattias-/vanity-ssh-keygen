package matcher

import "github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"

type Matcher interface {
	Name() string
	SetMatchString(string)
	Match(*key.SSHKey) bool
}
