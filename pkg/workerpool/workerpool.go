package workerpool

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Worker[R any] interface {
	Run()
	Count() int64
	SetResultChan(R)
}

type WorkerPool[R any] struct {
	Workers []Worker[R]
	Results R
	start   time.Time
}

type WorkerPoolStats struct {
	Workers int
	Count   int64
	Elapsed time.Duration
}

func (wps WorkerPoolStats) Log() {
	slog.Info("Tested keys",
		slog.Duration("time", wps.Elapsed),
		slog.Int64("tested", wps.Count),
		slog.Float64("kKeys/s", float64(wps.Count)/wps.Elapsed.Seconds()/1000),
	)
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
	meter := otel.Meter("keygen")
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
		slog.Warn("failed to initialize instrument", "error", err)
	}
}

func (wp *WorkerPool[R]) GetStats() *WorkerPoolStats {
	var sum int64
	for _, w := range wp.Workers {
		sum += w.Count()
	}
	return &WorkerPoolStats{
		Workers: len(wp.Workers),
		Count:   sum,
		Elapsed: time.Since(wp.start),
	}
}
