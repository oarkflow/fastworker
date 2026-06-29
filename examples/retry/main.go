package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	var attempts atomic.Int64
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 100})
	_ = p.Start()
	_ = p.SubmitFunc(func(ctx context.Context) error {
		if attempts.Add(1) < 3 {
			return errors.New("temporary")
		}
		fmt.Println("success on attempt", attempts.Load())
		return nil
	}, fastworker.JobOptions{MaxAttempts: 5, Backoff: fastworker.ConstantBackoff(10 * time.Millisecond)})
	time.Sleep(100 * time.Millisecond)
	_ = p.Stop(context.Background())
	fmt.Printf("retries=%d completed=%d\n", p.Stats().Retried, p.Stats().Completed)
}
