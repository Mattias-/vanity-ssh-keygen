package worker

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/matcher"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/ssh/key"
)

var testedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tested_keys_total",
		Help: "Number of tested keys.",
	},
	[]string{"matcher", "keygen"},
)

func init() {
	prometheus.MustRegister(testedTotal)
}

type worker struct {
	count        uint64
	matcher      matcher.Matcher
	keyGenerator func() key.SSHKey
}

func (w *worker) run(result chan *key.SSHKey) {
	m := testedTotal.WithLabelValues(w.matcher.Name(), "ed25519")
	var k key.SSHKey
	for {
		m.Inc()
		w.count += 1
		k = w.keyGenerator()
		if w.matcher.Match(&k) {
			// A result was found!
			break
		}
	}
	result <- &k
}

type WorkerPool struct {
	workers []*worker
	start   time.Time
	Results chan *key.SSHKey
}

func NewWorkerPool(instances int, matcher matcher.Matcher, keygen func() key.SSHKey) *WorkerPool {
	var workers []*worker
	for i := 0; i < instances; i++ {
		w := &worker{
			matcher:      matcher,
			keyGenerator: keygen,
		}
		workers = append(workers, w)
	}
	return &WorkerPool{
		workers: workers,
		start:   time.Now(),
		Results: make(chan *key.SSHKey),
	}
}

func (wp *WorkerPool) Start() {
	for _, w := range wp.workers {
		go w.run(wp.Results)
	}
}

type WorkerPoolStats struct {
	Workers int
	Count   uint64
	Elapsed time.Duration
}

func (wp *WorkerPool) GetStats() *WorkerPoolStats {
	var sum uint64
	for _, w := range wp.workers {
		sum += w.count
	}
	return &WorkerPoolStats{
		Workers: len(wp.workers),
		Count:   sum,
		Elapsed: time.Since(wp.start),
	}
}
