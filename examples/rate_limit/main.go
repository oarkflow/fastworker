package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 2, QueueSize: 100})

	// Default mode is RateLimitQueue: submit accepts quickly, workers process at
	// the configured pace in the background. Use Mode: RateLimitReject only when
	// you explicitly want admission errors.
	p.SetRateLimit("email", fastworker.RateLimit{Rate: 2, Burst: 1})
	_ = p.Start()
	defer p.Terminate()

	var done atomic.Int64
	for i := 0; i < 6; i++ {
		i := i
		err := p.SubmitFunc(func(context.Context) error {
			fmt.Printf("processed request=%d at=%s\n", i, time.Now().Format("15:04:05.000"))
			done.Add(1)
			return nil
		}, fastworker.JobOptions{RateLimitKey: "email"})
		fmt.Println("accepted", i, "err", err)
	}

	for done.Load() < 6 {
		time.Sleep(100 * time.Millisecond)
	}
}
