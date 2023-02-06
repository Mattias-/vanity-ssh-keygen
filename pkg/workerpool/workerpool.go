package workerpool

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var meter = global.Meter("keygen")

type Worker[R any] interface {
	Run()
	Count() uint64
	SetResultChan(R)
}

type WorkerPool[R any] struct {
	Workers []Worker[R]
	Results R
	start   time.Time
}

type WorkerPoolStats struct {
	Workers int
	Count   uint64
	Elapsed time.Duration
}

func (wp *WorkerPool[R]) Start() {
	wp.RegisterCounter()
	wp.start = time.Now()
	for _, w := range wp.Workers {
		w.SetResultChan(wp.Results)
		go w.Run()
	}
}

func (wp *WorkerPool[R]) RegisterCounter() {
	_, err := meter.Int64ObservableCounter(
		"keys.generated",
		instrument.WithDescription("Keys generated"),
		instrument.WithUnit("{keys}"),
		instrument.WithInt64Callback(func(ctx context.Context, o instrument.Int64Observer) error {
			var sum int64
			for _, w := range wp.Workers {
				sum += int64(w.Count())
			}
			o.Observe(sum)
			return nil
		}),
	)
	if err != nil {
		log.Fatalf("failed to initialize instrument: %v", err)
	}
}

func (wp *WorkerPool[R]) GetStats() *WorkerPoolStats {
	var sum uint64
	for _, w := range wp.Workers {
		sum += w.Count()
	}
	return &WorkerPoolStats{
		Workers: len(wp.Workers),
		Count:   sum,
		Elapsed: time.Since(wp.start),
	}
}
