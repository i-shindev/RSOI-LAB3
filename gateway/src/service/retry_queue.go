package service

import (
	ctx "context"
	"log/slog"
	"time"
)

type Job struct {
	name string
	work func(ctx ctx.Context) error
}

type retryQueue struct {
	retry          chan Job
	retry_duration time.Duration
}

type RetryQueue interface {
	ScheduleRetry(name string, work func(ctx ctx.Context) error)
	Run(ctx ctx.Context)
}

func NewRetryQueue(duration time.Duration) RetryQueue {
	return &retryQueue{make(chan Job), duration}
}

func (rq *retryQueue) ScheduleRetry(name string, work func(ctx ctx.Context) error) {
	time.AfterFunc(rq.retry_duration, func() { rq.retry <- Job{name, work} })
}

func (rq *retryQueue) Run(ctx ctx.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-rq.retry:

			err := job.work(ctx)
			if err == nil {
				slog.Info("retry succeeded", "job", job.name)
			} else {
				slog.Warn("retry failed with an error", "job", job.name)
			}
		}

	}
}
