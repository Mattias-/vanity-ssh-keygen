package ignorecase

import (
	"testing"
)

type mockSSHKey struct {
	pubkey []byte
}

func (m *mockSSHKey) SSHPubkey() []byte {
	return m.pubkey
}

func (m *mockSSHKey) SSHPrivkey() []byte {
	return nil
}

func (m *mockSSHKey) Generate() {}

func TestIgnoreCaseMatcher(t *testing.T) {
	m := New()
	m.SetMatchString("abc")

	testCases := []struct {
		pubkey string
		match  bool
	}{
		{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQABC", true},
		{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQabc", true},
		{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQXYZ", false},
	}

	for _, tc := range testCases {
		key := &mockSSHKey{pubkey: []byte(tc.pubkey)}
		if m.Match(key) != tc.match {
			t.Errorf("Expected match=%v for pubkey %s", tc.match, tc.pubkey)
		}
	}
}
