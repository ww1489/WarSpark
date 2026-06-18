package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

func TestImageSearchWorkerRunOnceProcessesJob(t *testing.T) {
	processor := &fakeImageSearchProcessor{
		job: imagesearchdomain.Job{
			ID:           "job_123",
			SearchStatus: imagesearchdomain.StatusMatched,
		},
	}
	worker := NewImageSearchWorker(processor, testLogger(), ImageSearchWorkerOptions{})

	processed, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if !processed {
		t.Fatal("expected processed=true")
	}
	if processor.calls != 1 {
		t.Fatalf("expected ProcessNext called once, got %d", processor.calls)
	}
}

func TestImageSearchWorkerRunOnceIgnoresNoPendingJob(t *testing.T) {
	processor := &fakeImageSearchProcessor{err: imagesearchdomain.ErrNoPendingJob}
	worker := NewImageSearchWorker(processor, testLogger(), ImageSearchWorkerOptions{})

	processed, err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if processed {
		t.Fatal("expected processed=false")
	}
	if processor.calls != 1 {
		t.Fatalf("expected ProcessNext called once, got %d", processor.calls)
	}
}

func TestImageSearchWorkerRunOnceReturnsUnexpectedError(t *testing.T) {
	expectedErr := errors.New("db unavailable")
	processor := &fakeImageSearchProcessor{err: expectedErr}
	worker := NewImageSearchWorker(processor, testLogger(), ImageSearchWorkerOptions{})

	processed, err := worker.RunOnce(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if processed {
		t.Fatal("expected processed=false")
	}
}

type fakeImageSearchProcessor struct {
	calls int
	job   imagesearchdomain.Job
	err   error
}

func (f *fakeImageSearchProcessor) ProcessNext(context.Context) (imagesearchdomain.Job, error) {
	f.calls++
	return f.job, f.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
