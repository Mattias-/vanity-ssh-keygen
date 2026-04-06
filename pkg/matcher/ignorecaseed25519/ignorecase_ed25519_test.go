package ignorecaseed25519

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

func TestIgnoreCaseEd25519Matcher(t *testing.T) {
	m := New()
	m.SetMatchString("abc")

	testCases := []struct {
		pubkey string
		match  bool
	}{
		// The matcher starts at index 37
		{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIABC", true},
		{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIabc", true},
		{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIXYZ", false},
	}

	for _, tc := range testCases {
		key := &mockSSHKey{pubkey: []byte(tc.pubkey)}
		if m.Match(key) != tc.match {
			t.Errorf("Expected match=%v for pubkey %s", tc.match, tc.pubkey)
		}
	}
}
