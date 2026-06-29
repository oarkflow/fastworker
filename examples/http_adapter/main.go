package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/oarkflow/fastworker"
)

func main() {
	p := fastworker.MustNewPool(fastworker.PresetAPIBuffer())
	_ = p.Start()
	defer p.Terminate()

	http.Handle("/ingest", fastworker.HTTPMiddleware(p, fastworker.HTTPOptions{
		Queue:        "api",
		RespondJobID: true,
		MaxBodySize:  1 << 20,
		Handler: func(ctx context.Context, req fastworker.HTTPRequest) error {
			fmt.Println("background request", req.Method, req.URL, len(req.Body))
			return nil
		},
	}))
	log.Println("POST http://localhost:8090/ingest")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
