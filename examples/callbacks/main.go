package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	pool := fastworker.MustNewPool(fastworker.WithWorkers(1))
	if err := pool.Start(); err != nil {
		panic(err)
	}

	attempts := 0
	cb := &fastworker.Callback{
		OnQueued: func(ctx context.Context, j fastworker.JobInfo) { fmt.Println("queued", j.ID) },
		OnStart:  func(ctx context.Context, j fastworker.JobInfo) { fmt.Println("started", j.ID, "attempt", j.Attempts) },
		OnProgress: func(ctx context.Context, j fastworker.JobInfo, p fastworker.Progress) {
			fmt.Println("progress", p.Percent, p.Message)
		},
		OnRetry: func(ctx context.Context, j fastworker.JobInfo, attempt int, err error) {
			fmt.Println("retry", attempt, err)
		},
		OnSuccess:    func(ctx context.Context, j fastworker.JobInfo) { fmt.Println("success", j.ID) },
		OnDeadLetter: func(ctx context.Context, j fastworker.JobInfo, err error) { fmt.Println("deadletter", j.ID, err) },
		OnFinally:    func(ctx context.Context, j fastworker.JobInfo, err error) { fmt.Println("finally", j.ID, err) },
	}

	_ = pool.SubmitFunc(func(ctx context.Context) error {
		attempts++
		fastworker.ReportProgress(ctx, attempts*50, "processing")
		if attempts == 1 {
			return fastworker.Retryable(errors.New("temporary outage"))
		}
		return nil
	}, fastworker.JobOptions{ID: "callback-demo", MaxAttempts: 2, Backoff: fastworker.ConstantBackoff(20 * time.Millisecond), Callback: cb})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = pool.WaitIdle(ctx)
	_ = pool.Shutdown(ctx)
}
