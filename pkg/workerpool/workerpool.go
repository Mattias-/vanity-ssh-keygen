package workerpool

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

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
	var meter = otel.Meter("keygen")
	_, err := meter.Int64ObservableCounter(
		"keys.generated",
		metric.WithDescription("Keys generated"),
		metric.WithUnit("{keys}"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
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
