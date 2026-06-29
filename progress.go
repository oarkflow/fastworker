package fastworker

import "context"

type Progress struct {
	Percent int    `json:"percent,omitempty"`
	Message string `json:"message,omitempty"`
}

type progressContextKey struct{}

func contextWithProgress(ctx context.Context, qj *queuedJob) context.Context {
	if qj == nil {
		return ctx
	}
	return context.WithValue(ctx, progressContextKey{}, qj)
}

func ReportProgress(ctx context.Context, percent int, message string) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	qj, _ := ctx.Value(progressContextKey{}).(*queuedJob)
	if qj == nil {
		return
	}
	qj.progress.Store(Progress{Percent: percent, Message: message})
}
