package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	var active atomic.Int64
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 8, QueueSize: 100})
	_ = p.Start()
	for i := 0; i < 5; i++ {
		_ = p.SubmitFunc(func(context.Context) error {
			n := active.Add(1)
			fmt.Println("active", n)
			time.Sleep(20 * time.Millisecond)
			active.Add(-1)
			return nil
		}, fastworker.JobOptions{ConcurrencyKey: "customer:1", MaxConcurrentPerKey: 1})
	}
	time.Sleep(200 * time.Millisecond)
	_ = p.Stop(context.Background())
}
