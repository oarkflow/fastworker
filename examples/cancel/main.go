package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 8, EnableJobTracking: true})
	_ = p.Start()
	f, _ := fastworker.SubmitResult[string](p, func(ctx context.Context) (string, error) {
		select {
		case <-time.After(time.Second):
			return "done", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}, fastworker.JobOptions{ID: "slow"})
	time.Sleep(20 * time.Millisecond)
	fmt.Println("cancelled", p.CancelJob("slow"))
	_, err := f.Get(context.Background())
	fmt.Println("future error", err)
	_ = p.Terminate()
}
