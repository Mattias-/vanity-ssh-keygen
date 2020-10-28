package matcher

import (
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

type ignorecaseEd25519Matcher struct {
	name        string
	matchString string
}

func NewIgnorecaseEd25519Matcher() *ignorecaseEd25519Matcher {
	return &ignorecaseEd25519Matcher{
		name: "ignorecase-ed25519",
	}
}

func (m *ignorecaseEd25519Matcher) Name() string {
	return m.name
}

func (m *ignorecaseEd25519Matcher) SetMatchString(matchString string) {
	m.matchString = matchString
}

func (m *ignorecaseEd25519Matcher) Match(s *key.SSHKey) bool {
	pubK := (*s).SSHPubkey()
	return strings.Contains(strings.ToLower(string(pubK[:37])), m.matchString)
}
