package fastworker

import "context"

type Callback struct {
	OnSuccess func(context.Context, JobInfo)
	OnError   func(context.Context, JobInfo, error)
}
