package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 100})
	_ = p.Start()
	start := time.Now()
	_ = p.ScheduleAfter("later", 100*time.Millisecond, fastworker.JobFunc(func(context.Context) error {
		fmt.Println("ran after", time.Since(start).Round(10*time.Millisecond))
		return nil
	}))
	time.Sleep(150 * time.Millisecond)
	_ = p.Stop(context.Background())
}
