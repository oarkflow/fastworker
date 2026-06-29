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
	_ = p.SubmitBytes([]byte("hello"), func(ctx context.Context, b []byte) error { fmt.Println(string(b)); return nil })
	_ = p.SubmitMap(map[string]any{"kind": "event"}, func(ctx context.Context, m map[string]any) error { fmt.Println(m["kind"]); return nil })
	_ = p.WaitIdle(context.Background())
}
