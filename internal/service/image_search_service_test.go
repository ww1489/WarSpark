package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

func TestImageSearchServiceCreateJobNormalizesAndStoresUpload(t *testing.T) {
	repository := &fakeImageSearchRepository{}
	storage := &fakeImageStorage{}
	service := NewImageSearchService(repository, storage)

	job, err := service.CreateJob(context.Background(), imagesearchdomain.UploadInput{
		FileName: "enemy-base.png",
		Size:     int64(len(testPNG(t, 512, 512))),
		Reader:   bytes.NewReader(testPNG(t, 512, 512)),
		ClientIP: "203.0.113.7",
		TargetContext: imagesearchdomain.TargetContext{
			WarTargetID: "target_123",
			EnemyName:   "No. 7",
		},
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if job.ID == "" {
		t.Fatal("expected generated job ID")
	}
	if job.SearchStatus != "created" {
		t.Fatalf("expected created status, got %q", job.SearchStatus)
	}
	if job.UploadedImage == nil {
		t.Fatal("expected uploaded image metadata")
	}
	if job.UploadedImage.Width != 512 || job.UploadedImage.Height != 512 {
		t.Fatalf("expected 512x512 dimensions, got %dx%d", job.UploadedImage.Width, job.UploadedImage.Height)
	}
	if !job.UploadedImage.Normalized {
		t.Fatal("expected normalized upload")
	}
	if job.UploadedImage.RawRetentionDays != 7 {
		t.Fatalf("expected raw retention of 7 days, got %d", job.UploadedImage.RawRetentionDays)
	}
	if len(storage.saved) == 0 || !bytes.HasPrefix(storage.saved, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatal("expected normalized PNG bytes to be saved")
	}
	if repository.created.UploadIPHash == "" {
		t.Fatal("expected anonymous upload IP hash")
	}
	if repository.created.TargetContext.WarTargetID != "target_123" {
		t.Fatalf("expected target context to be persisted, got %#v", repository.created.TargetContext)
	}
}

func TestImageSearchServiceCreateJobRejectsUnsupportedType(t *testing.T) {
	service := NewImageSearchService(&fakeImageSearchRepository{}, &fakeImageStorage{})

	_, err := service.CreateJob(context.Background(), imagesearchdomain.UploadInput{
		FileName: "base.txt",
		Size:     4,
		Reader:   bytes.NewReader([]byte("nope")),
	})
	if err == nil {
		t.Fatal("expected unsupported file error")
	}
	if got := imagesearchdomain.ErrorCode(err); got != "unsupported_file_type" {
		t.Fatalf("expected unsupported_file_type, got %q", got)
	}
}

func TestImageSearchServiceCreateJobRejectsSmallImage(t *testing.T) {
	data := testPNG(t, 320, 512)
	service := NewImageSearchService(&fakeImageSearchRepository{}, &fakeImageStorage{})

	_, err := service.CreateJob(context.Background(), imagesearchdomain.UploadInput{
		FileName: "small.png",
		Size:     int64(len(data)),
		Reader:   bytes.NewReader(data),
	})
	if err == nil {
		t.Fatal("expected image_too_small error")
	}
	if got := imagesearchdomain.ErrorCode(err); got != "image_too_small" {
		t.Fatalf("expected image_too_small, got %q", got)
	}
}

func TestImageSearchServiceProcessNextMatchesSameTHCandidates(t *testing.T) {
	expectedTH := 16
	repository := &fakeImageSearchRepository{
		claimedJob: imagesearchdomain.Job{
			ID:           "job_123",
			SearchStatus: imagesearchdomain.StatusProcessing,
			UploadedImage: &imagesearchdomain.UploadedImage{
				ID:     "img_123",
				Width:  1440,
				Height: 1440,
			},
			TargetContext: imagesearchdomain.TargetContext{ExpectedTH: &expectedTH},
		},
		candidates: []imagesearchdomain.CandidateLayout{
			{
				LayoutID:        "layout_123",
				MatchLevel:      "medium",
				ConfidenceScore: ptrFloat(0.6),
				ResultReason:    "same_th_public_layout",
			},
		},
	}
	service := NewImageSearchService(repository, &fakeImageStorage{})

	job, err := service.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext returned error: %v", err)
	}

	if repository.candidateFilter.THLevel == nil || *repository.candidateFilter.THLevel != 16 {
		t.Fatalf("expected candidate search by TH16, got %#v", repository.candidateFilter)
	}
	if repository.completed.SearchStatus != imagesearchdomain.StatusMatched {
		t.Fatalf("expected matched status, got %#v", repository.completed)
	}
	if repository.completed.DetectedTH == nil || *repository.completed.DetectedTH != 16 {
		t.Fatalf("expected detected TH16, got %#v", repository.completed)
	}
	if len(repository.completed.Candidates) != 1 || repository.completed.Candidates[0].LayoutID != "layout_123" {
		t.Fatalf("expected candidate to be persisted, got %#v", repository.completed.Candidates)
	}
	if job.SearchStatus != imagesearchdomain.StatusMatched {
		t.Fatalf("expected returned matched job, got %#v", job)
	}
}

func TestImageSearchServiceProcessNextMarksNoResultWithoutCandidates(t *testing.T) {
	repository := &fakeImageSearchRepository{
		claimedJob: imagesearchdomain.Job{
			ID:           "job_123",
			SearchStatus: imagesearchdomain.StatusProcessing,
			UploadedImage: &imagesearchdomain.UploadedImage{
				ID:     "img_123",
				Width:  512,
				Height: 512,
			},
		},
	}
	service := NewImageSearchService(repository, &fakeImageStorage{})

	job, err := service.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext returned error: %v", err)
	}

	if repository.completed.SearchStatus != imagesearchdomain.StatusNoResult {
		t.Fatalf("expected no_result status, got %#v", repository.completed)
	}
	if repository.completed.ScreenshotQuality != "acceptable" {
		t.Fatalf("expected acceptable quality, got %#v", repository.completed)
	}
	if len(repository.completed.Candidates) != 0 {
		t.Fatalf("expected no candidates, got %#v", repository.completed.Candidates)
	}
	if job.SearchStatus != imagesearchdomain.StatusNoResult {
		t.Fatalf("expected returned no_result job, got %#v", job)
	}
}

type fakeImageSearchRepository struct {
	created         imagesearchdomain.CreateJobInput
	claimedJob      imagesearchdomain.Job
	candidateFilter imagesearchdomain.CandidateFilter
	candidates      []imagesearchdomain.CandidateLayout
	completed       imagesearchdomain.CompleteJobInput
}

func (f *fakeImageSearchRepository) CreateJob(_ context.Context, input imagesearchdomain.CreateJobInput) (imagesearchdomain.Job, error) {
	f.created = input
	return imagesearchdomain.Job{
		ID:           input.JobID,
		SearchStatus: input.SearchStatus,
		UploadedImage: &imagesearchdomain.UploadedImage{
			ID:               input.ImageID,
			ImageURL:         input.ImageURL,
			Width:            input.Width,
			Height:           input.Height,
			Normalized:       input.Normalized,
			RawRetentionDays: input.RawRetentionDays,
		},
		TargetContext: input.TargetContext,
		CreatedAt:     input.CreatedAt,
	}, nil
}

func (f *fakeImageSearchRepository) GetJob(context.Context, string) (imagesearchdomain.Job, error) {
	return imagesearchdomain.Job{}, nil
}

func (f *fakeImageSearchRepository) GetResults(context.Context, string) (imagesearchdomain.Results, error) {
	return imagesearchdomain.Results{}, nil
}

func (f *fakeImageSearchRepository) RetryJob(context.Context, string) (imagesearchdomain.Job, error) {
	return imagesearchdomain.Job{}, nil
}

func (f *fakeImageSearchRepository) ClaimNextJob(context.Context) (imagesearchdomain.Job, error) {
	if f.claimedJob.ID == "" {
		return imagesearchdomain.Job{}, imagesearchdomain.ErrNoPendingJob
	}
	return f.claimedJob, nil
}

func (f *fakeImageSearchRepository) FindCandidateLayouts(_ context.Context, filter imagesearchdomain.CandidateFilter) ([]imagesearchdomain.CandidateLayout, error) {
	f.candidateFilter = filter
	return f.candidates, nil
}

func (f *fakeImageSearchRepository) CompleteJob(_ context.Context, input imagesearchdomain.CompleteJobInput) (imagesearchdomain.Job, error) {
	f.completed = input
	return imagesearchdomain.Job{
		ID:                input.JobID,
		SearchStatus:      input.SearchStatus,
		DetectedTH:        input.DetectedTH,
		ScreenshotQuality: input.ScreenshotQuality,
		BuildingsDetected: input.BuildingsDetected,
	}, nil
}

type fakeImageStorage struct {
	saved []byte
	key   string
}

func (f *fakeImageStorage) Save(_ context.Context, key string, reader io.Reader) (string, error) {
	f.key = key
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	f.saved = data
	return "/uploads/" + key, nil
}

func testPNG(t *testing.T, width int, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func ptrFloat(value float64) *float64 {
	return &value
}
