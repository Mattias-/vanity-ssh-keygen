package matcher

import (
	"log"
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

type lowercaseMatcher struct {
	matchString string
}

func NewLowercaseMatcher() *lowercaseMatcher {
	return &lowercaseMatcher{}
}

func (m *lowercaseMatcher) Name() string {
	return "lowercase"
}

func (m *lowercaseMatcher) SetMatchString(matchString string) {
	m.matchString = matchString
}

func (m *lowercaseMatcher) Match(s *key.SSHKey) bool {
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
