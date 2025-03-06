package matcher

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type Matcher interface {
	SetMatchString(string)
	Match(keygen.SSHKey) bool
}

var matchers = map[string]Matcher{}

func RegisterMatcher(name string, m Matcher) {
	matchers[name] = m
}

func Names() []string {
	keys := make([]string, 0, len(matchers))
	for k := range matchers {
		keys = append(keys, k)
	}
	return keys
}

func Get(name string) (Matcher, bool) {
	m, ok := matchers[name]
	return m, ok
}
