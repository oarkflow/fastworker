package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 16})
	_ = p.Start()
	var n atomic.Int64
	h, _ := p.ScheduleCron("heartbeat", "@every 100ms", fastworker.JobFunc(func(context.Context) error {
		fmt.Println("tick", n.Add(1))
		return nil
	}))
	time.Sleep(350 * time.Millisecond)
	h.Cancel()
	_ = p.Stop(context.Background())
}
