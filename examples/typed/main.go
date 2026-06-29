package main

import (
	"context"
	"fmt"

	"github.com/oarkflow/fastworker"
)

type Email struct{ To string }
type Receipt struct{ ID string }

func main() {
	p := fastworker.MustNewTyped[Email, Receipt](fastworker.Config{MinWorkers: 2}, func(ctx context.Context, e Email) (Receipt, error) {
		return Receipt{ID: "sent:" + e.To}, nil
	})
	_ = p.Start()
	f, _ := p.Submit(context.Background(), Email{To: "user@example.com"})
	r, _ := f.Get(context.Background())
	fmt.Println(r.ID)
	_ = p.Stop(context.Background())
}
