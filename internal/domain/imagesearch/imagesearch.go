package imagesearch

import (
	"errors"
	"io"
	"time"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
)

const (
	StatusCreated       = "created"
	StatusProcessing    = "processing"
	StatusMatched       = "matched"
	StatusLowConfidence = "low_confidence"
	StatusNoResult      = "no_result"
	StatusFailed        = "failed"

	ErrorUnsupportedFileType = "unsupported_file_type"
	ErrorFileTooLarge        = "file_too_large"
	ErrorImageTooSmall       = "image_too_small"
	ErrorRateLimited         = "rate_limited"
	ErrorImageUnreadable     = "image_unreadable"
	ErrorStorageFailed       = "storage_failed"
)

var ErrJobNotFound = errors.New("image search job not found")
var ErrNoPendingJob = errors.New("no pending image search job")

type UploadError struct {
	Code    string
	Message string
	Err     error
}

func (e UploadError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e UploadError) Unwrap() error {
	return e.Err
}

func NewUploadError(code string, message string) UploadError {
	return UploadError{Code: code, Message: message}
}

func WrapUploadError(code string, message string, err error) UploadError {
	return UploadError{Code: code, Message: message, Err: err}
}

func ErrorCode(err error) string {
	var uploadErr UploadError
	if errors.As(err, &uploadErr) {
		return uploadErr.Code
	}
	return ""
}

type TargetContext struct {
	WarTargetID   string `json:"war_target_id,omitempty"`
	EnemyPosition string `json:"enemy_position,omitempty"`
	EnemyName     string `json:"enemy_name,omitempty"`
	ExpectedTH    *int   `json:"expected_th,omitempty"`
}

type UploadInput struct {
	FileName      string
	Size          int64
	Reader        io.Reader
	ClientIP      string
	TargetContext TargetContext
}

type CreateJobInput struct {
	JobID                  string
	ImageID                string
	ImageURL               string
	Width                  int
	Height                 int
	Normalized             bool
	RawRetentionDays       int
	SearchStatus           string
	UploadIPHash           string
	OriginalRetentionUntil time.Time
	TargetContext          TargetContext
	CreatedAt              time.Time
}

type CandidateFilter struct {
	THLevel *int
	Limit   int
}

type CandidateLayout struct {
	LayoutID        string   `json:"layout_id"`
	MatchLevel      string   `json:"match_level"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`
	ResultReason    string   `json:"result_reason,omitempty"`
}

type CompleteJobInput struct {
	JobID             string
	SearchStatus      string
	DetectedTH        *int
	ScreenshotQuality string
	BuildingsDetected *int
	ErrorCode         string
	ErrorMessage      string
	Candidates        []CandidateLayout
}

type UploadedImage struct {
	ID               string `json:"image_id" db:"image_id"`
	ImageURL         string `json:"image_url" db:"image_url"`
	Width            int    `json:"width" db:"width"`
	Height           int    `json:"height" db:"height"`
	Normalized       bool   `json:"normalized" db:"-"`
	RawRetentionDays int    `json:"raw_retention_days" db:"-"`
}

type Job struct {
	ID                string         `json:"job_id" db:"id"`
	SearchStatus      string         `json:"search_status" db:"search_status"`
	UploadedImage     *UploadedImage `json:"uploaded_image,omitempty" db:"-"`
	DetectedTH        *int           `json:"detected_th,omitempty" db:"detected_th"`
	ScreenshotQuality string         `json:"screenshot_quality,omitempty" db:"screenshot_quality"`
	BuildingsDetected *int           `json:"buildings_detected,omitempty" db:"buildings_detected"`
	ErrorCode         string         `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage      string         `json:"error_message,omitempty" db:"error_message"`
	TargetContext     TargetContext  `json:"target_context,omitempty" db:"-"`
	ReviewQueueID     string         `json:"review_queue_id,omitempty" db:"-"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at,omitempty" db:"updated_at"`
}

type MatchSummary struct {
	CandidateCount int    `json:"candidate_count"`
	BestMatchLevel string `json:"best_match_level,omitempty"`
	LowConfidence  bool   `json:"low_confidence"`
}

type Results struct {
	Job            Job                       `json:"job"`
	MatchSummary   MatchSummary              `json:"match_summary"`
	Layouts        []layoutdomain.Card       `json:"layouts"`
	AttackVideos   []layoutdomain.VideoMatch `json:"attack_videos"`
	DefenseReplays []layoutdomain.VideoMatch `json:"defense_replays"`
	Fallback       any                       `json:"fallback"`
}
