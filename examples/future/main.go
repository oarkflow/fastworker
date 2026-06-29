package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 2, QueueSize: 100})
	_ = p.Start()
	f, _ := fastworker.SubmitResult[int](p, func(ctx context.Context) (int, error) { return 42, nil })
	v, err := f.Get(context.Background())
	fmt.Println(v, err)
	_ = p.Stop(context.Background())
}
