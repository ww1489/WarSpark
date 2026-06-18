package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

type ImageSearchProcessor interface {
	ProcessNext(ctx context.Context) (imagesearchdomain.Job, error)
}

type ImageSearchWorkerOptions struct {
	Interval time.Duration
}

type ImageSearchWorker struct {
	processor ImageSearchProcessor
	logger    *slog.Logger
	interval  time.Duration
}

func NewImageSearchWorker(processor ImageSearchProcessor, logger *slog.Logger, options ImageSearchWorkerOptions) *ImageSearchWorker {
	interval := options.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ImageSearchWorker{
		processor: processor,
		logger:    logger,
		interval:  interval,
	}
}

func (w *ImageSearchWorker) RunOnce(ctx context.Context) (bool, error) {
	job, err := w.processor.ProcessNext(ctx)
	if err != nil {
		if errors.Is(err, imagesearchdomain.ErrNoPendingJob) {
			return false, nil
		}
		return false, err
	}
	w.logger.Info("image search job processed", slog.String("job_id", job.ID), slog.String("status", job.SearchStatus))
	return true, nil
}

func (w *ImageSearchWorker) Start(ctx context.Context) {
	go w.loop(ctx)
}

func (w *ImageSearchWorker) loop(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		if _, err := w.RunOnce(ctx); err != nil {
			w.logger.Warn("image search worker run failed", slog.String("error", err.Error()))
		}

		select {
		case <-ctx.Done():
			w.logger.Info("image search worker stopped")
			return
		case <-ticker.C:
		}
	}
}
