package ignorecaseed25519

import (
	"bytes"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type ignorecaseEd25519Matcher struct {
	matchString []byte
}

func New() *ignorecaseEd25519Matcher {
	return &ignorecaseEd25519Matcher{}
}

func (m *ignorecaseEd25519Matcher) SetMatchString(matchString string) {
	m.matchString = []byte(matchString)
}

func (m *ignorecaseEd25519Matcher) Match(s keygen.SSHKey) bool {
	pubK := s.SSHPubkey()
	/*
	   The public key is 80 bytes long, the first 37 bytes are the key type and the key length, the last 43 bytes are the key itself
	   https://crypto.stackexchange.com/questions/44584/ed25519-ssh-public-key-is-always-80-characters-long
	   Example public key:
	   ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINyqBBX74InG199sb0Dg+5+vhuHbBimaTtiJb0+OGbzJ
	   |   Key type and length             || key                                     |
	*/
	return bytes.Contains(bytes.ToLower(pubK[37:]), m.matchString)
}
