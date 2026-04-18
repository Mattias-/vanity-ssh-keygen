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
	m.matchString = strings.ToLower(matchString)
}

func (m *ignorecaseMatcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	return containsCaseInsensitive(pubK, m.matchString)
}

func containsCaseInsensitive(b []byte, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(b)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c := b[i+j]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
