package matcher

import (
	"errors"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
)

type Matcher interface {
	Name() string
	SetMatchString(string)
	Match(*sshkey.SSHKey) bool
}

func MatcherList() mli {
	return []Matcher{
		NewIgnorecaseMatcher(),
		NewIgnorecaseEd25519Matcher(),
	}
}

type mli []Matcher

func (ml mli) Names() []string {
	var ms []string
	for _, m := range ml {
		ms = append(ms, m.Name())
	}
	return ms
}

func (ml mli) Get(name string) (Matcher, error) {
	for _, m := range ml {
		if name == m.Name() {
			return m, nil
		}
	}
	return nil, errors.New("Unknown item")
}
