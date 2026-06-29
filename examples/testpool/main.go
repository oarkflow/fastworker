package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.NewTestPool()
	defer p.Terminate()
	_ = p.SubmitFunc(func(context.Context) error { fmt.Println("test job"); return nil })
	_ = p.RunAll(context.Background())
}
