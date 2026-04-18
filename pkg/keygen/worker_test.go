package keygen

import (
	"context"
	"testing"
)

type mockWorkerKey struct {
	count int
	match int
}

func (m *mockWorkerKey) SSHPubkey() []byte  { return nil }
func (m *mockWorkerKey) SSHPrivkey() []byte { return nil }
func (m *mockWorkerKey) Generate() {
	m.count++
}

func TestWorker(t *testing.T) {
	results := make(chan SSHKey, 1)

	key := &mockWorkerKey{match: 5}

	w := &Worker{
		Matchfunc: func(k SSHKey) bool {
			return k.(*mockWorkerKey).count == k.(*mockWorkerKey).match
		},
		Keyfunc: func() SSHKey {
			return key
		},
	}
	w.SetResultChan(results)

	w.Run(context.Background())

	res := <-results
	if res == nil {
		t.Fatal("Expected result, got nil")
	}

	if w.Count() != 5 {
		t.Errorf("Expected count 5, got %d", w.Count())
	}

	if res.(*mockWorkerKey).count != 5 {
		t.Errorf("Expected key generate count 5, got %d", res.(*mockWorkerKey).count)
	}
}
