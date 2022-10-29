package worker

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
)

var meter = global.Meter("keygen")

type matcher interface {
	Name() string
	SetMatchString(string)
	Match(*sshkey.SSHKey) bool
}

type keygen interface {
	Name() string
	New() sshkey.SSHKey
}

type kgworker struct {
	count   uint64
	matcher matcher
	keygen  keygen
	results chan *sshkey.SSHKey
}

func (w *kgworker) run() {
	k := w.keygen.New()
	for {
		w.count += 1
		k.New()
		if w.matcher.Match(&k) {
			// A result was found!
			break
		}
	}
	w.results <- &k
}

type workerPool struct {
	workers []*kgworker
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
		w := &kgworker{
			matcher: matcher,
			keygen:  kg,
			results: wp.Results,
		}
		wp.workers = append(wp.workers, w)
	}

	counter, err := meter.AsyncInt64().Counter(
		"keys.generated",
		instrument.WithDescription("Keys generated"),
		instrument.WithUnit("{keys}"),
	)
	if err != nil {
		log.Fatalf("failed to initialize instrument: %v", err)
	}

	err = meter.RegisterCallback([]instrument.Asynchronous{counter},
		func(ctx context.Context) {
			var sum int64
			for _, w := range wp.workers {
				sum += int64(w.count)
			}
			counter.Observe(ctx, sum)
		})
	if err != nil {
		log.Fatalf("failed to register instrument callback: %v", err)
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
		sum += w.count
	}
	return &WorkerPoolStats{
		Workers: len(wp.workers),
		Count:   sum,
		Elapsed: time.Since(wp.start),
	}
}
