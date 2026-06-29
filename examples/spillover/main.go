package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(1), fastworker.WithQueueSize(1), fastworker.WithBackpressure(fastworker.BackpressureReject), fastworker.WithSpillover(".fastworker-spill", 1000))
	_ = p.Start()
	defer p.Terminate()

	for i := 0; i < 10; i++ {
		i := i
		_ = p.SubmitFunc(func(context.Context) error { time.Sleep(10 * time.Millisecond); fmt.Println("job", i); return nil })
	}
	_ = p.WaitIdle(context.Background())
	fmt.Println("spill depth", p.SpillDepth())
}
