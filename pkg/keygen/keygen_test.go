package keygen

import (
	"slices"
	"testing"
)

type mockKey struct{}

func (m *mockKey) SSHPubkey() []byte  { return nil }
func (m *mockKey) SSHPrivkey() []byte { return nil }
func (m *mockKey) Generate()          {}

func TestRegistry(t *testing.T) {
	name := "mock"
	keygenFunc := func() SSHKey { return &mockKey{} }

	RegisterKeygen(name, keygenFunc)

	names := Names()
	found := slices.Contains(names, name)
	if !found {
		t.Fatalf("Expected %s in names, but not found", name)
	}

	kg, ok := Get(name)
	if !ok {
		t.Fatalf("Expected to find keygen %s, but not found", name)
	}
	if kg == nil {
		t.Fatal("Keygen func is nil")
	}

	_, ok = Get("non-existent")
	if ok {
		t.Fatal("Expected not to find non-existent keygen")
	}
}
