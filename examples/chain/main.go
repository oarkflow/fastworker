package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(2))
	_ = p.Start()
	defer p.Terminate()
	_ = p.Chain().
		ThenFunc(func(context.Context) error { fmt.Println("validate"); return nil }).
		ThenFunc(func(context.Context) error { fmt.Println("process"); return nil }).
		ThenFunc(func(context.Context) error { fmt.Println("notify"); return nil }).
		Submit(context.Background())
	_ = p.WaitIdle(context.Background())
}
