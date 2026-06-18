package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
)

const imageSearchRawRetentionDays = 7

type ImageSearchRepository struct {
	db *sqlx.DB
}

func NewImageSearchRepository(db *sqlx.DB) *ImageSearchRepository {
	return &ImageSearchRepository{db: db}
}

func (r *ImageSearchRepository) CreateJob(ctx context.Context, input imagesearchdomain.CreateJobInput) (imagesearchdomain.Job, error) {
	targetContextJSON, err := encodeTargetContext(input.TargetContext)
	if err != nil {
		return imagesearchdomain.Job{}, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("begin image search job transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO layout_images (
  id, layout_id, image_url, source_type, source_url, width, height,
  image_role, review_status, quality_status
) VALUES (?, NULL, ?, 'user_upload', NULL, ?, ?, 'upload', 'not_reviewed', 'unknown')`,
		input.ImageID, input.ImageURL, input.Width, input.Height)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("insert uploaded image: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO image_search_jobs (
  id, uploaded_image_id, search_status, upload_ip_hash, original_retention_until,
  processing_mode, target_context, created_at, updated_at
) VALUES (?, ?, ?, NULLIF(?, ''), ?, 'auto', CAST(? AS JSON), ?, ?)`,
		input.JobID,
		input.ImageID,
		input.SearchStatus,
		input.UploadIPHash,
		input.OriginalRetentionUntil,
		targetContextJSON,
		input.CreatedAt,
		input.CreatedAt,
	)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("insert image search job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("commit image search job transaction: %w", err)
	}

	return r.GetJob(ctx, input.JobID)
}

func (r *ImageSearchRepository) GetJob(ctx context.Context, id string) (imagesearchdomain.Job, error) {
	return r.getJob(ctx, id)
}

func (r *ImageSearchRepository) GetResults(ctx context.Context, id string) (imagesearchdomain.Results, error) {
	job, err := r.getJob(ctx, id)
	if err != nil {
		return imagesearchdomain.Results{}, err
	}

	layouts, bestMatchLevel, err := r.resultLayouts(ctx, id)
	if err != nil {
		return imagesearchdomain.Results{}, err
	}

	attackVideos, defenseReplays, err := r.resultVideos(ctx, layouts)
	if err != nil {
		return imagesearchdomain.Results{}, err
	}

	return imagesearchdomain.Results{
		Job: job,
		MatchSummary: imagesearchdomain.MatchSummary{
			CandidateCount: len(layouts),
			BestMatchLevel: bestMatchLevel,
			LowConfidence:  job.SearchStatus == imagesearchdomain.StatusLowConfidence,
		},
		Layouts:        layouts,
		AttackVideos:   attackVideos,
		DefenseReplays: defenseReplays,
		Fallback:       nil,
	}, nil
}

func (r *ImageSearchRepository) RetryJob(ctx context.Context, id string) (imagesearchdomain.Job, error) {
	result, err := r.db.ExecContext(ctx, `
UPDATE image_search_jobs
SET search_status = 'created',
    error_code = NULL,
    error_message = NULL,
    updated_at = CURRENT_TIMESTAMP(3)
WHERE id = ? AND search_status IN ('failed', 'no_result', 'low_confidence')`, id)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("retry image search job: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("retry image search job rows affected: %w", err)
	}
	if affected == 0 {
		if _, err := r.getJob(ctx, id); err != nil {
			return imagesearchdomain.Job{}, err
		}
	}
	return r.getJob(ctx, id)
}

func (r *ImageSearchRepository) ClaimNextJob(ctx context.Context) (imagesearchdomain.Job, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("begin claim image search job: %w", err)
	}
	defer tx.Rollback()

	var id string
	if err := tx.GetContext(ctx, &id, `
SELECT id
FROM image_search_jobs
WHERE search_status = 'created'
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return imagesearchdomain.Job{}, imagesearchdomain.ErrNoPendingJob
		}
		return imagesearchdomain.Job{}, fmt.Errorf("select image search job to claim: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE image_search_jobs
SET search_status = 'processing',
    updated_at = CURRENT_TIMESTAMP(3)
WHERE id = ?`, id); err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("claim image search job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("commit claim image search job: %w", err)
	}
	return r.getJob(ctx, id)
}

func (r *ImageSearchRepository) FindCandidateLayouts(ctx context.Context, filter imagesearchdomain.CandidateFilter) ([]imagesearchdomain.CandidateLayout, error) {
	if filter.THLevel == nil {
		return []imagesearchdomain.CandidateLayout{}, nil
	}
	limit := filter.Limit
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	var ids []string
	if err := r.db.SelectContext(ctx, &ids, `
SELECT id
FROM base_layouts
WHERE th_level = ?
  AND visibility = 'public'
  AND review_status = 'reviewed'
  AND quality_status <> 'rejected'
ORDER BY updated_at DESC
LIMIT ?`, *filter.THLevel, limit); err != nil {
		return nil, fmt.Errorf("find candidate layouts: %w", err)
	}

	score := 0.6
	candidates := make([]imagesearchdomain.CandidateLayout, 0, len(ids))
	for _, id := range ids {
		candidates = append(candidates, imagesearchdomain.CandidateLayout{
			LayoutID:        id,
			MatchLevel:      "medium",
			ConfidenceScore: &score,
			ResultReason:    "same_th_public_layout",
		})
	}
	return candidates, nil
}

func (r *ImageSearchRepository) CompleteJob(ctx context.Context, input imagesearchdomain.CompleteJobInput) (imagesearchdomain.Job, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("begin complete image search job: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM image_search_results WHERE search_job_id = ?", input.JobID); err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("delete previous image search results: %w", err)
	}

	for index, candidate := range input.Candidates {
		_, err := tx.ExecContext(ctx, `
INSERT INTO image_search_results (
  id, search_job_id, layout_id, rank_order, match_level, confidence_score, result_reason
) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''))`,
			uuid.NewString(),
			input.JobID,
			candidate.LayoutID,
			index+1,
			candidate.MatchLevel,
			candidate.ConfidenceScore,
			candidate.ResultReason,
		)
		if err != nil {
			return imagesearchdomain.Job{}, fmt.Errorf("insert image search result: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
UPDATE image_search_jobs
SET search_status = ?,
    detected_th = ?,
    screenshot_quality = NULLIF(?, ''),
    buildings_detected = ?,
    error_code = NULLIF(?, ''),
    error_message = NULLIF(?, ''),
    updated_at = CURRENT_TIMESTAMP(3)
WHERE id = ?`,
		input.SearchStatus,
		input.DetectedTH,
		input.ScreenshotQuality,
		input.BuildingsDetected,
		input.ErrorCode,
		input.ErrorMessage,
		input.JobID,
	)
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("complete image search job: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("complete image search job rows affected: %w", err)
	}
	if affected == 0 {
		return imagesearchdomain.Job{}, imagesearchdomain.ErrJobNotFound
	}

	if err := tx.Commit(); err != nil {
		return imagesearchdomain.Job{}, fmt.Errorf("commit complete image search job: %w", err)
	}
	return r.getJob(ctx, input.JobID)
}

func (r *ImageSearchRepository) getJob(ctx context.Context, id string) (imagesearchdomain.Job, error) {
	var row imageSearchJobRow
	if err := r.db.GetContext(ctx, &row, `
SELECT
  j.id,
  j.search_status,
  j.detected_th,
  j.screenshot_quality,
  j.buildings_detected,
  j.error_code,
  j.error_message,
  COALESCE(CAST(j.target_context AS CHAR), '{}') AS target_context_json,
  j.created_at,
  j.updated_at,
  i.id AS image_id,
  i.image_url,
  COALESCE(i.width, 0) AS width,
  COALESCE(i.height, 0) AS height
FROM image_search_jobs j
LEFT JOIN layout_images i ON i.id = j.uploaded_image_id
WHERE j.id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return imagesearchdomain.Job{}, imagesearchdomain.ErrJobNotFound
		}
		return imagesearchdomain.Job{}, fmt.Errorf("get image search job: %w", err)
	}
	return row.job()
}

func (r *ImageSearchRepository) resultLayouts(ctx context.Context, jobID string) ([]layoutdomain.Card, string, error) {
	var rows []layoutCardRow
	if err := r.db.SelectContext(ctx, &rows, `
SELECT
  bl.id,
  bl.title,
  bl.th_level,
  bl.layout_type,
  COALESCE(CAST(bl.style_tags AS CHAR), '[]') AS style_tags_json,
  COALESCE((
    SELECT li.image_url
    FROM layout_images li
    WHERE li.layout_id = bl.id AND li.image_role IN ('primary', 'candidate', 'reference')
    ORDER BY FIELD(li.image_role, 'primary', 'candidate', 'reference'), li.created_at DESC
    LIMIT 1
  ), '') AS primary_image_url,
  bl.source_type,
  bl.review_status,
  bl.quality_status,
  COALESCE((
    SELECT ll.link_status
    FROM layout_links ll
    WHERE ll.layout_id = bl.id
    ORDER BY FIELD(ll.link_status, 'active', 'unverified', 'broken'), ll.created_at DESC
    LIMIT 1
  ), 'missing') AS link_status,
  bl.updated_at
FROM image_search_results isr
JOIN base_layouts bl ON bl.id = isr.layout_id
WHERE isr.search_job_id = ?
ORDER BY isr.rank_order ASC`, jobID); err != nil {
		return nil, "", fmt.Errorf("list image search result layouts: %w", err)
	}

	cards := make([]layoutdomain.Card, 0, len(rows))
	for _, row := range rows {
		card, err := row.card()
		if err != nil {
			return nil, "", err
		}
		cards = append(cards, card)
	}

	var bestMatchLevel string
	if err := r.db.GetContext(ctx, &bestMatchLevel, `
SELECT COALESCE(match_level, '')
FROM image_search_results
WHERE search_job_id = ?
ORDER BY rank_order ASC
LIMIT 1`, jobID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("get best image search match level: %w", err)
	}

	return cards, bestMatchLevel, nil
}

func (r *ImageSearchRepository) resultVideos(ctx context.Context, layouts []layoutdomain.Card) ([]layoutdomain.VideoMatch, []layoutdomain.VideoMatch, error) {
	if len(layouts) == 0 {
		return []layoutdomain.VideoMatch{}, []layoutdomain.VideoMatch{}, nil
	}

	attackByID := make(map[string]layoutdomain.VideoMatch)
	defenseByID := make(map[string]layoutdomain.VideoMatch)
	layoutRepository := &LayoutRepository{db: r.db}
	for _, card := range layouts {
		matches, err := layoutRepository.videoMatches(ctx, card.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, match := range matches {
			switch match.MatchGroup {
			case "attack_video":
				attackByID[match.MatchID] = match
			case "defense_replay":
				defenseByID[match.MatchID] = match
			}
		}
	}

	return sortedVideoMatches(attackByID), sortedVideoMatches(defenseByID), nil
}

func sortedVideoMatches(values map[string]layoutdomain.VideoMatch) []layoutdomain.VideoMatch {
	items := make([]layoutdomain.VideoMatch, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	sort.Slice(items, func(i int, j int) bool {
		if items[i].ConfidenceScore != nil && items[j].ConfidenceScore != nil {
			return *items[i].ConfidenceScore > *items[j].ConfidenceScore
		}
		return items[i].MatchID < items[j].MatchID
	})
	return items
}

type imageSearchJobRow struct {
	ID                string         `db:"id"`
	SearchStatus      string         `db:"search_status"`
	DetectedTH        sql.NullInt64  `db:"detected_th"`
	ScreenshotQuality sql.NullString `db:"screenshot_quality"`
	BuildingsDetected sql.NullInt64  `db:"buildings_detected"`
	ErrorCode         sql.NullString `db:"error_code"`
	ErrorMessage      sql.NullString `db:"error_message"`
	TargetContextJSON string         `db:"target_context_json"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	ImageID           sql.NullString `db:"image_id"`
	ImageURL          sql.NullString `db:"image_url"`
	Width             int            `db:"width"`
	Height            int            `db:"height"`
}

func (r imageSearchJobRow) job() (imagesearchdomain.Job, error) {
	targetContext, err := decodeTargetContext(r.TargetContextJSON)
	if err != nil {
		return imagesearchdomain.Job{}, err
	}

	job := imagesearchdomain.Job{
		ID:            r.ID,
		SearchStatus:  r.SearchStatus,
		TargetContext: targetContext,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if r.DetectedTH.Valid {
		value := int(r.DetectedTH.Int64)
		job.DetectedTH = &value
	}
	if r.ScreenshotQuality.Valid {
		job.ScreenshotQuality = r.ScreenshotQuality.String
	}
	if r.BuildingsDetected.Valid {
		value := int(r.BuildingsDetected.Int64)
		job.BuildingsDetected = &value
	}
	if r.ErrorCode.Valid {
		job.ErrorCode = r.ErrorCode.String
	}
	if r.ErrorMessage.Valid {
		job.ErrorMessage = r.ErrorMessage.String
	}
	if r.ImageID.Valid {
		job.UploadedImage = &imagesearchdomain.UploadedImage{
			ID:               r.ImageID.String,
			ImageURL:         r.ImageURL.String,
			Width:            r.Width,
			Height:           r.Height,
			Normalized:       true,
			RawRetentionDays: imageSearchRawRetentionDays,
		}
	}
	return job, nil
}

func encodeTargetContext(value imagesearchdomain.TargetContext) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode target context: %w", err)
	}
	return string(encoded), nil
}

func decodeTargetContext(value string) (imagesearchdomain.TargetContext, error) {
	if value == "" {
		return imagesearchdomain.TargetContext{}, nil
	}
	var result imagesearchdomain.TargetContext
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return imagesearchdomain.TargetContext{}, fmt.Errorf("decode target context: %w", err)
	}
	return result, nil
}
