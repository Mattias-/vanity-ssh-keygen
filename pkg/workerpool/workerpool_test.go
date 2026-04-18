package workerpool

import (
	"context"
	"testing"
	"time"
)

type mockWorker struct {
	count      int64
	resultChan chan int
}

func (m *mockWorker) Run(ctx context.Context) {
	m.count++
	if m.resultChan != nil {
		m.resultChan <- 1
	}
}

func (m *mockWorker) Count() int64 {
	return m.count
}

func (m *mockWorker) SetResultChan(c chan int) {
	m.resultChan = c
}

func TestWorkerPool(t *testing.T) {
	resChan := make(chan int, 1)
	w := &mockWorker{}
	wp := &WorkerPool[chan int]{
		Workers: []Worker[chan int]{w},
		Results: resChan,
	}

	wp.Start(context.Background())

	select {
	case <-resChan:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for worker to run")
	}

	stats := wp.GetStats()
	if stats.Workers != 1 {
		t.Errorf("Expected 1 worker, got %d", stats.Workers)
	}
	if stats.Count != 1 {
		t.Errorf("Expected count 1, got %d", stats.Count)
	}

	stats.Log()
}
