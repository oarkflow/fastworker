package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 2, MaxWorkers: 8, QueueSize: 1000})
	_ = p.Start()
	_ = p.Pause()
	fmt.Println("paused", p.Stats().Paused)
	for i := 0; i < 5; i++ {
		i := i
		_ = p.SubmitFunc(func(ctx context.Context) error { fmt.Println("running", i); return nil })
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Println("completed while paused", p.Stats().Completed)
	_ = p.Resume()
	fmt.Println("resumed")
	time.Sleep(100 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = p.Stop(ctx)
	fmt.Printf("stopped completed=%d\n", p.Stats().Completed)

	p2 := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 10})
	_ = p2.Start()
	_ = p2.SubmitFunc(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
	_ = p2.Terminate()
	fmt.Println("terminated", p2.Stats().Terminated)
}
