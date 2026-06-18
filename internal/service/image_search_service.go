package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

const (
	maxUploadBytes      = 10 * 1024 * 1024
	minUploadSidePixels = 512
	maxUploadSidePixels = 4096
	rawRetentionDays    = 7
)

type ImageSearchRepository interface {
	CreateJob(ctx context.Context, input imagesearchdomain.CreateJobInput) (imagesearchdomain.Job, error)
	GetJob(ctx context.Context, id string) (imagesearchdomain.Job, error)
	GetResults(ctx context.Context, id string) (imagesearchdomain.Results, error)
	RetryJob(ctx context.Context, id string) (imagesearchdomain.Job, error)
	ClaimNextJob(ctx context.Context) (imagesearchdomain.Job, error)
	FindCandidateLayouts(ctx context.Context, filter imagesearchdomain.CandidateFilter) ([]imagesearchdomain.CandidateLayout, error)
	CompleteJob(ctx context.Context, input imagesearchdomain.CompleteJobInput) (imagesearchdomain.Job, error)
}

type ImageSearchService struct {
	repository ImageSearchRepository
	storage    ImageStorage
	now        func() time.Time
}

func NewImageSearchService(repository ImageSearchRepository, storage ImageStorage) *ImageSearchService {
	return &ImageSearchService{
		repository: repository,
		storage:    storage,
		now:        time.Now,
	}
}

func (s *ImageSearchService) CreateJob(ctx context.Context, input imagesearchdomain.UploadInput) (imagesearchdomain.Job, error) {
	if !isSupportedImageFile(input.FileName) {
		return imagesearchdomain.Job{}, imagesearchdomain.NewUploadError(imagesearchdomain.ErrorUnsupportedFileType, "unsupported file type")
	}
	if input.Size > maxUploadBytes {
		return imagesearchdomain.Job{}, imagesearchdomain.NewUploadError(imagesearchdomain.ErrorFileTooLarge, "file too large")
	}

	uploadBytes, err := readUpload(input.Reader)
	if err != nil {
		return imagesearchdomain.Job{}, err
	}

	decoded, _, err := image.Decode(bytes.NewReader(uploadBytes))
	if err != nil {
		return imagesearchdomain.Job{}, imagesearchdomain.WrapUploadError(imagesearchdomain.ErrorImageUnreadable, "image unreadable", err)
	}

	width := decoded.Bounds().Dx()
	height := decoded.Bounds().Dy()
	if width < minUploadSidePixels || height < minUploadSidePixels {
		return imagesearchdomain.Job{}, imagesearchdomain.NewUploadError(imagesearchdomain.ErrorImageTooSmall, "image too small")
	}

	normalized := resizeIfNeeded(decoded)
	var output bytes.Buffer
	if err := png.Encode(&output, normalized); err != nil {
		return imagesearchdomain.Job{}, imagesearchdomain.WrapUploadError(imagesearchdomain.ErrorImageUnreadable, "image unreadable", err)
	}

	jobID := uuid.NewString()
	imageID := uuid.NewString()
	key := filepath.ToSlash(filepath.Join("image-search", jobID, imageID+".png"))
	imageURL, err := s.storage.Save(ctx, key, bytes.NewReader(output.Bytes()))
	if err != nil {
		return imagesearchdomain.Job{}, imagesearchdomain.WrapUploadError(imagesearchdomain.ErrorStorageFailed, "storage failed", err)
	}

	now := s.now().UTC()
	return s.repository.CreateJob(ctx, imagesearchdomain.CreateJobInput{
		JobID:                  jobID,
		ImageID:                imageID,
		ImageURL:               imageURL,
		Width:                  normalized.Bounds().Dx(),
		Height:                 normalized.Bounds().Dy(),
		Normalized:             true,
		RawRetentionDays:       rawRetentionDays,
		SearchStatus:           imagesearchdomain.StatusCreated,
		UploadIPHash:           hashUploadIP(input.ClientIP),
		OriginalRetentionUntil: now.Add(rawRetentionDays * 24 * time.Hour),
		TargetContext:          input.TargetContext,
		CreatedAt:              now,
	})
}

func (s *ImageSearchService) GetJob(ctx context.Context, id string) (imagesearchdomain.Job, error) {
	return s.repository.GetJob(ctx, id)
}

func (s *ImageSearchService) GetResults(ctx context.Context, id string) (imagesearchdomain.Results, error) {
	return s.repository.GetResults(ctx, id)
}

func (s *ImageSearchService) RetryJob(ctx context.Context, id string) (imagesearchdomain.Job, error) {
	return s.repository.RetryJob(ctx, id)
}

func (s *ImageSearchService) ProcessNext(ctx context.Context) (imagesearchdomain.Job, error) {
	job, err := s.repository.ClaimNextJob(ctx)
	if err != nil {
		return imagesearchdomain.Job{}, err
	}

	detectedTH := job.DetectedTH
	if detectedTH == nil && job.TargetContext.ExpectedTH != nil {
		value := *job.TargetContext.ExpectedTH
		detectedTH = &value
	}

	screenshotQuality := classifyScreenshotQuality(job.UploadedImage)
	var candidates []imagesearchdomain.CandidateLayout
	if detectedTH != nil {
		candidates, err = s.repository.FindCandidateLayouts(ctx, imagesearchdomain.CandidateFilter{
			THLevel: detectedTH,
			Limit:   5,
		})
		if err != nil {
			return s.repository.CompleteJob(ctx, imagesearchdomain.CompleteJobInput{
				JobID:             job.ID,
				SearchStatus:      imagesearchdomain.StatusFailed,
				DetectedTH:        detectedTH,
				ScreenshotQuality: screenshotQuality,
				ErrorCode:         "matcher_unavailable",
				ErrorMessage:      err.Error(),
			})
		}
	}

	status := imagesearchdomain.StatusNoResult
	if len(candidates) > 0 {
		status = imagesearchdomain.StatusMatched
	}

	return s.repository.CompleteJob(ctx, imagesearchdomain.CompleteJobInput{
		JobID:             job.ID,
		SearchStatus:      status,
		DetectedTH:        detectedTH,
		ScreenshotQuality: screenshotQuality,
		Candidates:        candidates,
	})
}

func readUpload(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, imagesearchdomain.NewUploadError(imagesearchdomain.ErrorImageUnreadable, "image unreadable")
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxUploadBytes+1))
	if err != nil {
		return nil, imagesearchdomain.WrapUploadError(imagesearchdomain.ErrorImageUnreadable, "image unreadable", err)
	}
	if len(data) > maxUploadBytes {
		return nil, imagesearchdomain.NewUploadError(imagesearchdomain.ErrorFileTooLarge, "file too large")
	}
	return data, nil
}

func isSupportedImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func resizeIfNeeded(img image.Image) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	maxSide := width
	if height > maxSide {
		maxSide = height
	}
	if maxSide <= maxUploadSidePixels {
		return img
	}

	ratio := float64(maxUploadSidePixels) / float64(maxSide)
	targetWidth := int(float64(width) * ratio)
	targetHeight := int(float64(height) * ratio)
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

func hashUploadIP(ip string) string {
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}

func classifyScreenshotQuality(image *imagesearchdomain.UploadedImage) string {
	if image == nil {
		return "unknown"
	}
	if image.Width >= 1024 && image.Height >= 1024 {
		return "good"
	}
	return "acceptable"
}
