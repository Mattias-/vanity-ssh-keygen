package ignorecaseed25519

import (
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type ignorecaseEd25519Matcher struct {
	matchString []byte
}

func New() *ignorecaseEd25519Matcher {
	return &ignorecaseEd25519Matcher{}
}

func (m *ignorecaseEd25519Matcher) SetMatchString(matchString string) {
	m.matchString = []byte(strings.ToLower(matchString))
}

func (m *ignorecaseEd25519Matcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	if len(pubK) < 37 {
		return false
	}
	// The public key is 81 bytes long. The base64 part starts at index 37.
	return containsCaseInsensitive(pubK[37:], m.matchString)
}

func containsCaseInsensitive(b, substr []byte) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(b)-len(substr); i++ {
		match := true
		for j := range substr {
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
