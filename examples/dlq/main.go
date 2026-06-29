package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 16})
	_ = p.Start()
	_ = p.SubmitFunc(func(context.Context) error { return errors.New("always fails") }, fastworker.JobOptions{ID: "bad", MaxAttempts: 1})
	_ = p.WaitIdle(context.Background())
	fmt.Println("dlq", len(p.DeadLetters()))
	fmt.Println("purged", p.PurgeDeadLetters())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = p.Stop(ctx)
}
