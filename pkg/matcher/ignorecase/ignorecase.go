package ignorecase

import (
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type ignorecaseMatcher struct {
	matchString string
}

func New() *ignorecaseMatcher {
	return &ignorecaseMatcher{}
}

func (m *ignorecaseMatcher) SetMatchString(matchString string) {
	m.matchString = matchString
}

func (m *ignorecaseMatcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	return strings.Contains(strings.ToLower(string(pubK)), m.matchString)
}
