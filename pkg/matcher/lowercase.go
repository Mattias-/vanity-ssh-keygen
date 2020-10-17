package matcher

import (
	"log"
	"strings"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

type LowercaseMatcher struct {
	MatchString string
}

func (m LowercaseMatcher) Match(s *key.SSHKey) bool {
	if s == nil {
		return false
	}
	pubK, err := (*s).SSHPubkey()
	if err != nil {
		log.Println(err)
		return false
	}
	return strings.Contains(strings.ToLower(string(pubK)), m.MatchString)
}
