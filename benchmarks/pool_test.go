package benchmarks

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/oarkflow/fastworker"
)

func BenchmarkSubmitRun(b *testing.B) {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 16, MaxWorkers: 16, QueueSize: b.N, Backpressure: fastworker.BackpressureBlock})
	_ = p.Start()
	var done atomic.Int64
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.SubmitFunc(func(context.Context) error { done.Add(1); return nil })
	}
	for done.Load() < int64(b.N) {
	}
	b.StopTimer()
	_ = p.Stop(context.Background())
}

type atomicJob struct{ done *atomic.Int64 }

func (j atomicJob) Run(context.Context) error { j.done.Add(1); return nil }

func BenchmarkSubmitRunReusableJob(b *testing.B) {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 16, MaxWorkers: 16, QueueSize: b.N, Backpressure: fastworker.BackpressureBlock})
	_ = p.Start()
	var done atomic.Int64
	job := atomicJob{done: &done}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Submit(job)
	}
	for done.Load() < int64(b.N) {
	}
	b.StopTimer()
	_ = p.Stop(context.Background())
}
