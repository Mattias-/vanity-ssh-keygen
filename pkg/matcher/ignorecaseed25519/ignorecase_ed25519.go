package ignorecaseed25519

import (
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type ignorecaseEd25519Matcher struct {
	matchString string
}

func New() *ignorecaseEd25519Matcher {
	return &ignorecaseEd25519Matcher{}
}

func (m *ignorecaseEd25519Matcher) SetMatchString(matchString string) {
	m.matchString = matchString
}

func (m *ignorecaseEd25519Matcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	return strings.Contains(strings.ToLower(string(pubK[37:])), m.matchString)
}
