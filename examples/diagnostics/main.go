package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.WithWorkers(1), fastworker.WithQueueSize(4))
	_ = p.Start()
	defer p.Terminate()

	for i := 0; i < 4; i++ {
		_ = p.SubmitFunc(func(context.Context) error { time.Sleep(50 * time.Millisecond); return nil })
	}
	fmt.Println(p.ExplainSaturation())
	fmt.Println(p.Dump())
}
