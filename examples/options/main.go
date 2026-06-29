package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(
		fastworker.PresetBackgroundQueue(),
		fastworker.WithWorkers(4),
		fastworker.WithQueueSize(10_000),
		fastworker.WithRateLimit("api", fastworker.RateLimit{Rate: 100, Burst: 200}),
		fastworker.WithDefaultOptions(fastworker.JobOptions{MaxAttempts: 3, Timeout: 5 * time.Second}),
	)
	_ = p.Start()
	defer p.Terminate()

	_ = p.SubmitFunc(func(context.Context) error { fmt.Println("configured with functional options"); return nil }, fastworker.JobOptions{RateLimitKey: "api"})
	_ = p.WaitIdle(context.Background())
}
