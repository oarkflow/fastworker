package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 4, QueueSize: 100})
	_ = p.Start()
	res, _ := fastworker.Map[int, int](context.Background(), p, []int{1, 2, 3, 4}, func(ctx context.Context, v int) (int, error) { return v * v, nil })
	fmt.Println(res.Values)
	_ = p.Stop(context.Background())
}
