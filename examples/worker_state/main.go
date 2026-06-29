package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(1), fastworker.WithWorkerInit(func(ctx context.Context, w *fastworker.Worker) error {
		w.Set("client", "reusable-client")
		fmt.Println("worker init", w.ID)
		return nil
	}), fastworker.WithWorkerClose(func(ctx context.Context, w *fastworker.Worker) error {
		fmt.Println("worker close", w.ID)
		return nil
	}))
	_ = p.Start()
	_ = p.SubmitFunc(func(context.Context) error {
		fmt.Println("job uses worker-owned resources via init hook pattern")
		return nil
	})
	_ = p.WaitIdle(context.Background())
	_ = p.Stop(context.Background())
}
