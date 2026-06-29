package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	lc := fastworker.NewLifecycle()
	lc.On(func(ctx context.Context, e fastworker.Event) {
		if e.ErrorText != "" {
			fmt.Printf("event=%s job=%s attempt=%d err=%s\n", e.Type, e.Job.ID, e.Attempt, e.ErrorText)
			return
		}
		if e.Job.ID != "" {
			fmt.Printf("event=%s job=%s attempt=%d progress=%d%%\n", e.Type, e.Job.ID, e.Attempt, e.Job.Progress.Percent)
			return
		}
		fmt.Printf("event=%s state=%s worker=%d\n", e.Type, e.State, e.WorkerID)
	},
		fastworker.EventPoolStarted,
		fastworker.EventWorkerStarted,
		fastworker.EventJobSubmitted,
		fastworker.EventJobStarted,
		fastworker.EventJobProgress,
		fastworker.EventJobSucceeded,
		fastworker.EventPoolStopped,
	)

	pool := fastworker.MustNewPool(
		fastworker.WithWorkers(2),
		fastworker.WithLifecycle(lc),
		fastworker.WithJobTracking(true),
	)
	if err := pool.Start(); err != nil {
		panic(err)
	}

	_ = pool.SubmitFunc(func(ctx context.Context) error {
		fastworker.ReportProgress(ctx, 25, "loaded input")
		time.Sleep(10 * time.Millisecond)
		fastworker.ReportProgress(ctx, 100, "done")
		return nil
	}, fastworker.JobOptions{ID: "lifecycle-demo"})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = pool.WaitIdle(ctx)
	_ = pool.Shutdown(ctx)
}
