package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	dir, _ := os.MkdirTemp("", "fastworker-store-*")
	store, err := fastworker.NewFileStore(dir)
	if err != nil {
		panic(err)
	}
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 1, QueueSize: 16}, fastworker.WithStore(store))
	_ = p.Start()
	_ = p.SubmitFunc(func(context.Context) error { return errors.New("boom") }, fastworker.JobOptions{ID: "audit-job", MaxAttempts: 1})
	time.Sleep(30 * time.Millisecond)
	_ = p.Stop(context.Background())
	fmt.Println("audit files written to", dir)
}
