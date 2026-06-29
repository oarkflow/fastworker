package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(1), fastworker.WithJobTracking(true))
	_ = p.Start()
	defer p.Terminate()

	_ = p.SubmitFunc(func(ctx context.Context) error {
		for i := 0; i <= 100; i += 25 {
			fastworker.ReportProgress(ctx, i, "working")
			time.Sleep(20 * time.Millisecond)
		}
		return nil
	}, fastworker.JobOptions{ID: "progress-demo"})

	for {
		info, ok := p.InspectJob("progress-demo")
		if ok {
			fmt.Printf("progress: %+v\n", info.Progress)
		}
		if !ok || info.Progress.Percent == 100 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	_ = p.WaitIdle(context.Background())
}
