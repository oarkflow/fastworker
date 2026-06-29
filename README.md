# fastworker

`fastworker` is a self-contained, high-performance Go worker pool for modern services that need fast concurrent execution with strong lifecycle control and reliable job handling.

It is intentionally dependency-free and focuses on a fast in-memory execution engine with production controls.

## Features

- Pause, resume, graceful stop/drain, immediate terminate
- Fixed and autoscaling worker pools
- Idle worker reaping down to `MinWorkers`
- Bounded priority queue with backpressure policies
- Priority jobs
- Delayed and scheduled jobs
- Retry with constant/exponential backoff and jitter
- Dead-letter capture, replay, and purge
- Futures/results that correctly wait for retry completion
- Typed generic worker pools
- Batch map execution
- Per-key concurrency limits
- Token-bucket rate limiting
- Panic recovery with stack traces
- Hooks and middleware
- Timeout/logger/metadata middleware helpers
- Runtime stats and health-style state helpers
- Fast MPMC ring queue implementation for custom hot paths
- Zero third-party dependencies
- Examples, tests, and benchmark scaffold

## Quick start

```bash
cd fastworker
go test ./...
go run ./examples/basic
```

## Basic usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/oarkflow/fastworker"
)

func main() {
    p := fastworker.MustNew(fastworker.Config{
        MinWorkers: 4,
        MaxWorkers: 32,
        QueueSize:  100_000,
    })

    _ = p.Start()

    _ = p.SubmitFunc(func(ctx context.Context) error {
        fmt.Println("work")
        return nil
    })

    _ = p.WaitIdle(context.Background())
    _ = p.Stop(context.Background())
}
```

## Lifecycle controls

Pause prevents workers from starting more queued jobs while still accepting submissions:

```go
_ = p.Pause()
_ = p.Resume()
```

Graceful stop/drain rejects new jobs, drains accepted jobs, and waits for workers:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
_ = p.Stop(ctx)
```

Terminate cancels the pool context immediately. Running jobs receive cancellation through their context:

```go
_ = p.Terminate()
_ = p.Wait(context.Background())
```

Operational helpers:

```go
p.State().String()
p.IsRunning()
p.IsPaused()
p.QueueDepth()
p.WorkerCount()
p.WaitIdle(ctx)
p.Drain(ctx)
```

## Futures

```go
future, _ := fastworker.SubmitResult[int](p, func(ctx context.Context) (int, error) {
    return 42, nil
})

value, err := future.Get(context.Background())
```

Result futures do not complete on an intermediate retry failure. They complete only after final success, permanent failure, retry exhaustion, cancellation, or timeout.

## Typed pool

```go
type Email struct{ To string }
type Receipt struct{ ID string }

pool := fastworker.MustNewTyped[Email, Receipt](fastworker.Config{MinWorkers: 4},
    func(ctx context.Context, e Email) (Receipt, error) {
        return Receipt{ID: "sent:" + e.To}, nil
    },
)

_ = pool.Start()
f, _ := pool.Submit(context.Background(), Email{To: "user@example.com"})
r, _ := f.Get(context.Background())
fmt.Println(r.ID)
```

## Retry and DLQ

```go
_ = p.SubmitFunc(sendWebhook, fastworker.JobOptions{
    MaxAttempts: 5,
    Backoff: fastworker.ExponentialBackoff{
        Initial: 100 * time.Millisecond,
        Max:     10 * time.Second,
        Factor:  2,
        Jitter:  true,
    },
})
```

Return `fastworker.Permanent(err)` to prevent retries for known permanent failure:

```go
return fastworker.Permanent(err)
```

Inspect, replay, or purge DLQ jobs:

```go
dlq := p.DeadLetters()
replayed, err := p.ReplayDeadLetters()
purged := p.PurgeDeadLetters()
```

## Priority jobs

Higher priority runs first when jobs are ready at the same time.

```go
_ = p.SubmitFunc(job, fastworker.JobOptions{Priority: fastworker.PriorityCritical})
```

## Delayed and scheduled jobs

```go
_ = p.ScheduleAfter("cleanup", time.Minute, fastworker.JobFunc(cleanup))
_ = p.ScheduleAt("report", time.Now().Add(time.Hour), fastworker.JobFunc(report))
handle := p.ScheduleEvery("heartbeat", 5*time.Second, fastworker.JobFunc(heartbeat))
handle.Cancel()
```

The delayed queue wakes when an earlier job is submitted, so a long-delayed job will not block newly submitted immediate work.

## Rate limiting without rejecting requests

The default limiter mode is `RateLimitQueue`. This means `Submit` accepts the request into the bounded queue immediately and workers smooth execution in the background. Use this for HTTP/API ingestion where callers should not receive rate-limit errors just because downstream processing is paced.

```go
p.SetRateLimit("email", fastworker.RateLimit{Rate: 500, Burst: 1000}) // Mode defaults to RateLimitQueue

_ = p.SubmitFunc(sendEmail, fastworker.JobOptions{
    RateLimitKey: "email",
})
```

For arbitrary request payloads, use `SubmitAny`:

```go
type Request struct { ID string }

_ = p.SubmitAny(Request{ID: "req-1"}, func(ctx context.Context, payload any) error {
    req := payload.(Request)
    return process(req)
}, fastworker.JobOptions{RateLimitKey: "email"})
```

Traditional admission rejection is still available when explicitly requested:

```go
p.SetRateLimit("strict", fastworker.RateLimit{
    Rate: 100,
    Burst: 100,
    Mode: fastworker.RateLimitReject,
})
```

`RateLimitBlock` is also available for legacy callers that prefer `Submit` to wait for a token, but queue smoothing is the recommended production default.

## Per-key concurrency

```go
_ = p.SubmitFunc(processCustomer, fastworker.JobOptions{
    ConcurrencyKey:      "customer:123",
    MaxConcurrentPerKey: 1,
})
```

## Batch map

```go
result, err := fastworker.Map[int, int](ctx, p, []int{1,2,3}, func(ctx context.Context, n int) (int, error) {
    return n * n, nil
})
fmt.Println(result.Values, err)
```

## Hooks and middleware

```go
p := fastworker.MustNew(fastworker.Config{MinWorkers: 4},
    fastworker.WithHooks(fastworker.Hooks{
        OnJobError: func(ctx context.Context, opts fastworker.JobOptions, attempt int, err error) {
            log.Println("job failed", opts.ID, attempt, err)
        },
    }),
    fastworker.WithMiddleware(
        fastworker.Timeout(5*time.Second),
        fastworker.Logger("pool"),
        fastworker.WithMetadata(map[string]string{"service": "billing"}),
    ),
)
```

## Backpressure

```go
fastworker.BackpressureBlock
fastworker.BackpressureReject
fastworker.BackpressureDropOldest
fastworker.BackpressureDropNewest
```

`BackpressureBlock` waits until space is available. `SubmitTimeout` limits how long producers wait.

## Examples

```bash
go run ./examples/basic
go run ./examples/control
go run ./examples/future
go run ./examples/typed
go run ./examples/retry
go run ./examples/dlq
go run ./examples/priority
go run ./examples/delayed
go run ./examples/rate_limit
go run ./examples/concurrency
go run ./examples/batch
go run ./examples/middleware
```

## Benchmarks

```bash
go test -bench=. -benchmem ./...
```

## Package layout

```txt
fastworker/
  pool.go           core worker pool
  control.go        state, drain, wait, idle helpers
  job.go            job model and options
  future.go         futures/result support
  typed.go          generic typed pool
  batch.go          batch map helper
  scheduler.go      delayed and periodic scheduling
  backoff.go        retry backoff policies
  retry.go          retryable/permanent error helpers
  dlq.go            DLQ replay and purge
  rate_limiter.go   token bucket and keyed concurrency
  middleware.go     timeout/logger/metadata middleware
  hooks.go          lifecycle and job hooks
  metrics.go        stats counters
  queue/priority.go priority + delayed queue
  queue/ring.go     bounded MPMC ring
  examples/         runnable examples
  benchmarks/       benchmark scaffold
```

## Newly Added Production Controls

This archive also includes the extended operational layer:

- `CancelJob(id)` for queued/running job cancellation.
- `InspectJob(id)` for point-in-time job state lookup.
- `ActiveJobs()` for runtime job snapshots.
- `CancelAll(ctx)` for emergency queue/run cancellation.
- `AdminServer` with `/healthz`, `/readyz`, `/stats`, `/metrics`, `/pause`, `/resume`, `/terminate`, `/jobs`, `/jobs/{id}`, `/jobs/{id}/cancel`, `/dlq`, `/dlq/replay`, and `/dlq/purge`.
- Prometheus text exporter via `pool.Prometheus()`.
- `Store` interface and `FileStore` JSONL audit backend for job transitions and DLQ records.
- Cron-like scheduling with `ScheduleCron`, supporting `@every 5s`, `*/10 * * * * *` for second steps, and `*/2 * * * *` for minute steps.
- Extra runnable examples: `admin`, `store`, `cron`, and `cancel`.

### Admin API example

```bash
cd examples/admin
go run .
```

In another shell:

```bash
curl -H 'Authorization: Bearer dev-token' http://localhost:8090/stats
curl -H 'Authorization: Bearer dev-token' http://localhost:8090/metrics
curl -X POST -H 'Authorization: Bearer dev-token' http://localhost:8090/pause
curl -X POST -H 'Authorization: Bearer dev-token' http://localhost:8090/resume
```

### FileStore audit example

```go
store, err := fastworker.NewFileStore(".fastworker-data")
if err != nil {
    panic(err)
}

pool := fastworker.MustNew(
    fastworker.Config{MinWorkers: 4, QueueSize: 10000},
    fastworker.WithStore(store),
)
```

This writes append-only `jobs.jsonl` and `dlq.jsonl` records. It intentionally stores audit/state records instead of trying to serialize arbitrary Go closures.

### Cancellation and inspection

```go
future, _ := fastworker.SubmitResult[string](pool, func(ctx context.Context) (string, error) {
    select {
    case <-time.After(time.Second):
        return "done", nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}, fastworker.JobOptions{ID: "slow-job"})

info, ok := pool.InspectJob("slow-job")
_ = info
_ = ok

pool.CancelJob("slow-job")
_, err := future.Get(context.Background())
```

### Cron scheduling

```go
handle, err := pool.ScheduleCron("heartbeat", "@every 10s", fastworker.JobFunc(func(ctx context.Context) error {
    return nil
}))
if err != nil {
    panic(err)
}
defer handle.Cancel()
```

## Performance mode

The default hot path is optimized for simple immediate jobs:

- no automatic job ID generation
- no inspection-map writes unless `EnableJobTracking` is enabled
- no per-job context allocation unless the job needs cancellation/future/tracking
- no noop audit-store calls
- no heap/interface boxing for the priority queue
- immediate jobs use an O(1) bounded FIFO fast path
- advanced jobs still use the priority/delayed/retry path

For maximum throughput, submit a reusable `Job` value instead of allocating a new closure on every iteration:

```go
type MyJob struct{}
func (MyJob) Run(ctx context.Context) error { return nil }

job := MyJob{}
for i := 0; i < n; i++ {
    _ = p.Submit(job)
}
```

Enable tracking only when you need admin job inspection/cancellation by ID:

```go
p := fastworker.MustNew(fastworker.Config{
    MinWorkers:        16,
    MaxWorkers:        16,
    QueueSize:         1_000_000,
    EnableJobTracking: true,
})
```

The included benchmark now has both forms:

```bash
go test ./benchmarks -bench=. -benchmem
```

On the validation host, the closure benchmark dropped from the reported ~`2265 ns/op`, `720 B/op`, `10 allocs/op` to about `1000–1100 ns/op`, `16 B/op`, `1 alloc/op`. The reusable job benchmark runs around `890–960 ns/op`, `0 B/op`, `0 allocs/op`.

---

## Usability Layer Added

The library now includes a higher-level production API on top of the low-allocation worker core.

### Functional Options

```go
p := fastworker.MustNewPool(
    fastworker.PresetBackgroundQueue(),
    fastworker.WithWorkers(4),
    fastworker.WithQueueSize(100_000),
    fastworker.WithRateLimit("email", fastworker.RateLimit{Rate: 100, Burst: 500}),
    fastworker.WithDefaultOptions(fastworker.JobOptions{MaxAttempts: 3, Timeout: 5*time.Second}),
)
```

### Named Queues / Queue Policies

```go
p := fastworker.MustNewPool(
    fastworker.WithQueuePolicy("email", fastworker.QueuePolicy{
        RateLimitKey: "email",
        MaxAttempts:  3,
    }),
    fastworker.WithRateLimit("email", fastworker.RateLimit{Rate: 25, Burst: 50}),
)

email := p.Queue("email")
_ = email.SubmitFunc(func(ctx context.Context) error {
    return sendEmail(ctx)
})
```

Queues can be paused and resumed independently:

```go
p.PauseQueue("reports")
p.ResumeQueue("reports")
```

### Queue-first Rate Limiting

The default rate limiter mode is still queue-first: jobs are accepted quickly and execution is smoothed in the background.

```go
p.SetRateLimit("sms", fastworker.RateLimit{
    Rate:  100,
    Burst: 200,
    Mode:  fastworker.RateLimitQueue,
})
```

Use `RateLimitReject` only when admission must fail immediately.

### HTTP Async Adapter

```go
http.Handle("/ingest", fastworker.HTTPMiddleware(p, fastworker.HTTPOptions{
    Queue:        "api",
    RespondJobID: true,
    MaxBodySize:  1 << 20,
    Handler: func(ctx context.Context, req fastworker.HTTPRequest) error {
        return process(req.Body)
    },
}))
```

The adapter returns `202 Accepted` and processes the captured request in the background.

### Request Adapters

```go
_ = p.SubmitBytes(data, handleBytes)
_ = p.SubmitMap(map[string]any{"event": "created"}, handleMap)
_ = p.SubmitStruct(payload, handleAny)
_ = fastworker.SubmitChannel(p, ctx, events, handleEvent)
```

### Job Builder

```go
err := p.JobFunc(sendReport).
    Queue("reports").
    Priority(fastworker.PriorityHigh).
    Timeout(30*time.Second).
    Retry(3).
    RateLimit("reports").
    ConcurrencyKey("tenant:acme", 2).
    Metadata("tenant", "acme").
    Submit(ctx)
```

### Idempotency Convenience

```go
_ = p.Once("invoice:123", job)
_ = p.OncePer("email:user@example.com", time.Hour, job)
_ = p.Replace("sync:user:123", job)
_ = p.JoinExisting("report:monthly", job)
```

### Progress Reporting

```go
_ = p.SubmitFunc(func(ctx context.Context) error {
    fastworker.ReportProgress(ctx, 40, "uploading")
    fastworker.ReportProgress(ctx, 80, "processing")
    return nil
}, fastworker.JobOptions{ID: "job-1"})

info, _ := p.InspectJob("job-1")
fmt.Println(info.Progress)
```

Enable `WithJobTracking(true)` when you want live inspection.

### Callback Hooks Per Job

```go
_ = p.SubmitFunc(work, fastworker.JobOptions{
    Callback: &fastworker.Callback{
        OnSuccess: func(ctx context.Context, info fastworker.JobInfo) {},
        OnError:   func(ctx context.Context, info fastworker.JobInfo, err error) {},
    },
})
```

### Chain / WhenAll

```go
_ = p.Chain().
    ThenFunc(validate).
    ThenFunc(process).
    ThenFunc(notify).
    Submit(ctx)

_ = p.WhenAll(jobA, jobB, jobC).Then(finalJob).Submit(ctx)
```

### Spillover for Bursts

```go
p := fastworker.MustNewPool(
    fastworker.WithQueueSize(10_000),
    fastworker.WithBackpressure(fastworker.BackpressureReject),
    fastworker.WithSpillover("/var/lib/fastworker/spill", 1_000_000),
)
```

Spillover keeps function jobs in process memory while appending job audit snapshots to disk. It is designed to absorb bursts without blocking producers. Arbitrary Go closures are not restart-serializable after process crash; use a durable queue backend or payload codec for crash-replay semantics.

### Worker Init / Close

```go
p := fastworker.MustNewPool(
    fastworker.WithWorkerInit(func(ctx context.Context, w *fastworker.Worker) error {
        w.Set("client", newClient())
        return nil
    }),
    fastworker.WithWorkerClose(func(ctx context.Context, w *fastworker.Worker) error {
        return nil
    }),
)
```

### Diagnostics

```go
fmt.Println(p.ExplainSaturation())
fmt.Println(p.Dump())
diag := p.Diagnose()
```

### Config File

```yaml
name: fastworker
workers:
  min: 4
  max: 64
queue:
  size: 100000
  backpressure: reject
tracking: true
```

```go
p, err := fastworker.FromConfigFile("fastworker.yaml")
```

### Runtime Reconfiguration

```go
p.UpdateRateLimit("email", fastworker.RateLimit{Rate: 200, Burst: 500})
p.Resize(32)
p.UpdateQueueWeight("reports", fastworker.PriorityHigh)
p.PauseQueue("reports")
p.ResumeQueue("reports")
```

### Signal Helpers

```go
_ = fastworker.RunUntilSignal(p)
// or
_ = fastworker.ShutdownOnSignal(p, 30*time.Second)
```

### Testing Helpers

```go
p := fastworker.NewTestPool()
defer p.Terminate()
_ = p.SubmitFunc(testJob)
_ = p.RunAll(context.Background())
```

### New Examples

Additional examples are available under:

```txt
examples/options
examples/queues
examples/http_adapter
examples/builder
examples/progress
examples/diagnostics
examples/config
examples/spillover
examples/chain
examples/adapters
examples/worker_state
examples/testpool
```

### Performance Notes

The fastest path is still the immediate reusable `Job` path with no IDs, no tracking, no middleware, no store, no futures, no queue policy, and no metadata.

On the validation host:

```txt
BenchmarkSubmitRun                ~680 ns/op   16 B/op   1 alloc/op
BenchmarkSubmitRunReusableJob     ~609 ns/op    0 B/op   0 allocs/op
```

Use `PresetLowLatency` or `PresetHighThroughput` for hot paths, and enable the usability features only where they are needed.
