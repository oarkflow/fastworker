package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 4, MaxWorkers: 16, QueueSize: 10000})
	_ = p.Start()
	for i := 0; i < 10; i++ {
		i := i
		_ = p.SubmitFunc(func(ctx context.Context) error {
			fmt.Println("job", i)
			return nil
		})
	}
	_ = p.Stop(context.Background())
	fmt.Printf("completed=%d uptime=%s\n", p.Stats().Completed, p.Stats().Uptime.Truncate(time.Millisecond))
}
