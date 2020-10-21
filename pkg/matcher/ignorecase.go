package matcher

import (
	"log"
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
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

func (m *ignorecaseMatcher) Match(s *key.SSHKey) bool {
	if s == nil {
		return false
	}
	pubK, err := (*s).SSHPubkey()
	if err != nil {
		log.Println(err)
		return false
	}
	return strings.Contains(strings.ToLower(string(pubK)), m.matchString)
}
