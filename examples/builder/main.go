package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(2))
	_ = p.Start()
	defer p.Terminate()

	_ = p.JobFunc(func(context.Context) error { fmt.Println("built job"); return nil }).
		Queue("reports").
		Priority(fastworker.PriorityHigh).
		Timeout(time.Second).
		Retry(2).
		Metadata("tenant", "acme").
		Submit(context.Background())

	_ = p.WaitIdle(context.Background())
}
