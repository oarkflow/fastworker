package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNew(fastworker.Config{MinWorkers: 2, MaxWorkers: 8, QueueSize: 1024, Name: "workerpool", EnableJobTracking: true})
	_ = p.Start()
	srv := fastworker.AdminServer{Pool: p, Token: "dev-token"}.ListenAndServe(":8090")
	defer srv.Shutdown(context.Background())

	for i := 0; i < 10; i++ {
		i := i
		_ = p.SubmitFunc(func(context.Context) error { fmt.Println("job", i); time.Sleep(20 * time.Millisecond); return nil }, fastworker.JobOptions{ID: fmt.Sprintf("job-%d", i)})
	}
	fmt.Println("admin: curl -H 'Authorization: Bearer dev-token' localhost:8090/stats")
	_ = p.WaitIdle(context.Background())
	_ = p.Stop(context.Background())
}
