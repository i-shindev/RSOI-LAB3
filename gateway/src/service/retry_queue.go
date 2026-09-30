package service

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	name string
	work func(ctx context.Context) error
}

type RetryQueue struct {
	retry          chan Job
	retry_duration time.Duration
}

func NewRetryQueue(duration time.Duration) *RetryQueue {
	return &RetryQueue{make(chan Job), duration}
}

func (rq *RetryQueue) ScheduleRetry(name string, work func(ctx context.Context) error) {
	time.AfterFunc(rq.retry_duration, func() { rq.retry <- Job{name, work} })
}

func (rq *RetryQueue) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-rq.retry:

			err := job.work(ctx)
			if err == nil {
				slog.Info("retry succeeded", "job", job.name)
			} else {
				slog.Warn("retry failed with an error", "job", job.name, "error", err)
			}
		}

	}
}
