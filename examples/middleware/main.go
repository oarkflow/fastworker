package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 16},
		fastworker.WithMiddleware(fastworker.Timeout(time.Second), fastworker.Logger("worker"), fastworker.WithMetadata(map[string]string{"service": "billing"})),
	)
	_ = p.Start()
	_ = p.SubmitFunc(func(ctx context.Context) error {
		fmt.Println("metadata", fastworker.MetadataFromContext(ctx)["service"])
		return nil
	})
	_ = p.WaitIdle(context.Background())
	_ = p.Stop(context.Background())
}
