package matcher

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type Matcher interface {
	SetMatchString(string)
	Match(keygen.SSHKey) bool
}

type namedMatcher struct {
	name    string
	matcher Matcher
}

var matchers = []namedMatcher{}

func RegisterMatcher(name string, m Matcher) {
	matchers = append(matchers, namedMatcher{name, m})
}

func Names() []string {
	names := make([]string, 0, len(matchers))
	for _, k := range matchers {
		names = append(names, k.name)
	}
	return names
}

func Get(name string) (Matcher, bool) {
	for _, m := range matchers {
		if m.name == name {
			return m.matcher, true
		}
	}
	return nil, false
}
