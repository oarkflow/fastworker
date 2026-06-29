package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(
		fastworker.WithWorkers(2),
		fastworker.WithQueuePolicy("email", fastworker.QueuePolicy{RateLimitKey: "email", MaxAttempts: 3}),
		fastworker.WithRateLimit("email", fastworker.RateLimit{Rate: 5, Burst: 1}),
	)
	_ = p.Start()
	defer p.Terminate()

	email := p.Queue("email")
	for i := 0; i < 5; i++ {
		i := i
		_ = email.SubmitFunc(func(context.Context) error { fmt.Println("email", i); return nil })
	}
	_ = p.WaitIdle(context.Background())
	fmt.Println("done", time.Now().Format(time.Kitchen))
}
