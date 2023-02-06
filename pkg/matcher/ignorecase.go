package matcher

import (
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type ignorecaseMatcher struct {
	name        string
	matchString string
}

func NewIgnorecaseMatcher() *ignorecaseMatcher {
	return &ignorecaseMatcher{
		name: "ignorecase",
	}
}

func (m *ignorecaseMatcher) Name() string {
	return m.name
}

func (m *ignorecaseMatcher) SetMatchString(matchString string) {
	m.matchString = matchString
}

func (m *ignorecaseMatcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	return strings.Contains(strings.ToLower(string(pubK)), m.matchString)
}
