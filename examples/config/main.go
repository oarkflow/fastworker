package main

import (
	"context"
	"fmt"
	"os"

	"github.com/oarkflow/fastworker"
)

func main() {
	cfg := []byte("name: config-demo\nworkers:\n  min: 2\n  max: 4\nqueue:\n  size: 1000\n  backpressure: reject\ntracking: true\n")
	_ = os.WriteFile("fastworker.yaml", cfg, 0o644)
	p, err := fastworker.FromConfigFile("fastworker.yaml")
	if err != nil {
		panic(err)
	}
	_ = p.Start()
	defer p.Terminate()
	_ = p.SubmitFunc(func(context.Context) error { fmt.Println("loaded from config"); return nil })
	_ = p.WaitIdle(context.Background())
}
