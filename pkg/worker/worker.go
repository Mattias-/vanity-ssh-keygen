package worker

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
)

type matcher interface {
	Name() string
	SetMatchString(string)
	Match(*sshkey.SSHKey) bool
}

type keygen interface {
	Name() string
	New() sshkey.SSHKey
}

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

type worker interface {
	run()
	getCount() uint64
}

type kgworker struct {
	count   uint64
	matcher matcher
	keygen  keygen
	results chan *sshkey.SSHKey
}

func (w *kgworker) run() {
	m := testedTotal.WithLabelValues(w.matcher.Name(), w.keygen.Name())
	k := w.keygen.New()
	for {
		m.Inc()
		w.count += 1
		k.New()
		if w.matcher.Match(&k) {
			// A result was found!
			break
		}
	}
	w.results <- &k
}

func (w *kgworker) getCount() uint64 {
	return w.count
}

type workerPool struct {
	workers []worker
	start   time.Time
	Results chan *sshkey.SSHKey
}

type WorkerPoolStats struct {
	Workers int
	Count   uint64
	Elapsed time.Duration
}

func NewWorkerPool(instances int, matcher matcher, kg keygen) *workerPool {
	wp := &workerPool{
		Results: make(chan *sshkey.SSHKey),
	}
	for i := 0; i < instances; i++ {
		w := kgworker{
			matcher: matcher,
			keygen:  kg,
			results: wp.Results,
		}
		wp.workers = append(wp.workers, &w)
	}
	return wp
}

func (wp *workerPool) Start() {
	wp.start = time.Now()
	for _, w := range wp.workers {
		go w.run()
	}
}

func (wp *workerPool) GetStats() *WorkerPoolStats {
	var sum uint64
	for _, w := range wp.workers {
		sum += w.getCount()
	}
	return &WorkerPoolStats{
		Workers: len(wp.workers),
		Count:   sum,
		Elapsed: time.Since(wp.start),
	}
}
