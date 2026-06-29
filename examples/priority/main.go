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
	_ = p.Pause()
	_ = p.SubmitFunc(func(context.Context) error { fmt.Println("low"); return nil }, fastworker.JobOptions{Priority: fastworker.PriorityLow})
	_ = p.SubmitFunc(func(context.Context) error { fmt.Println("critical"); return nil }, fastworker.JobOptions{Priority: fastworker.PriorityCritical})
	_ = p.Resume()
	time.Sleep(50 * time.Millisecond)
	_ = p.Stop(context.Background())
}
