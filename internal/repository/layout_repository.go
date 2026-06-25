package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LayoutRepository struct {
	db *sqlx.DB
}

func NewLayoutRepository(db *sqlx.DB) *LayoutRepository {
	return &LayoutRepository{db: db}
}

func (r *LayoutRepository) List(ctx context.Context, filter layoutdomain.ListFilter, pagination utils.Pagination) (layoutdomain.ListResult, error) {
	where, args := layoutWhereClause(filter)

	var total int
	countQuery := "SELECT COUNT(*) FROM base_layouts bl " + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return layoutdomain.ListResult{}, fmt.Errorf("count layouts: %w", err)
	}

	query := `
SELECT
	bl.id,
	bl.title,
	bl.th_level,
	bl.layout_type,
	COALESCE(CAST(bl.style_tags AS CHAR), '[]') AS style_tags_json,
	COALESCE((
		SELECT li.image_url
		FROM layout_images li
		WHERE li.layout_id = bl.id AND li.image_role = 'primary'
		ORDER BY li.created_at DESC
		LIMIT 1
	), '') AS primary_image_url,
	bl.source_type,
	bl.review_status,
	bl.quality_status,
	COALESCE((
		SELECT ll.link_status
		FROM layout_links ll
		WHERE ll.layout_id = bl.id
		ORDER BY CASE ll.link_status
			WHEN 'active' THEN 1
			WHEN 'unverified' THEN 2
			WHEN 'broken' THEN 3
			ELSE 4
		END, ll.created_at DESC
		LIMIT 1
	), 'missing') AS link_status,
	bl.updated_at
FROM base_layouts bl
` + where + `
ORDER BY bl.updated_at DESC
LIMIT ? OFFSET ?`

	queryArgs := append(append([]any{}, args...), pagination.PageSize, pagination.Offset)
	var rows []layoutCardRow
	if err := r.db.SelectContext(ctx, &rows, query, queryArgs...); err != nil {
		return layoutdomain.ListResult{}, fmt.Errorf("list layouts: %w", err)
	}

	items := make([]layoutdomain.Card, 0, len(rows))
	for _, row := range rows {
		card, err := row.card()
		if err != nil {
			return layoutdomain.ListResult{}, err
		}
		items = append(items, card)
	}

	return layoutdomain.ListResult{Items: items, Total: total}, nil
}

func (r *LayoutRepository) Get(ctx context.Context, id string) (layoutdomain.Detail, error) {
	var row layoutDetailRow
	if err := r.db.GetContext(ctx, &row, `
SELECT
	id,
	title,
	th_level,
	layout_type,
	COALESCE(CAST(style_tags AS CHAR), '[]') AS style_tags_json,
	source_type,
	COALESCE(source_url, '') AS source_url,
	review_status,
	quality_status
FROM base_layouts
WHERE id = ? AND visibility = 'public'`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return layoutdomain.Detail{}, layoutdomain.ErrLayoutNotFound
		}
		return layoutdomain.Detail{}, fmt.Errorf("get layout: %w", err)
	}

	detail, err := row.detail()
	if err != nil {
		return layoutdomain.Detail{}, err
	}

	if err := r.db.SelectContext(ctx, &detail.Images, `
SELECT id, image_url, source_type, COALESCE(source_url, '') AS source_url, width, height, image_role, review_status, quality_status
FROM layout_images
WHERE layout_id = ?
ORDER BY CASE image_role WHEN 'primary' THEN 1 WHEN 'candidate' THEN 2 ELSE 3 END, created_at DESC`, id); err != nil {
		return layoutdomain.Detail{}, fmt.Errorf("list layout images: %w", err)
	}

	if err := r.db.SelectContext(ctx, &detail.Links, `
SELECT id, link_type, url, link_status, source_type, COALESCE(source_url, '') AS source_url, last_checked_at, COALESCE(last_check_error, '') AS last_check_error
FROM layout_links
WHERE layout_id = ?
ORDER BY CASE link_status WHEN 'active' THEN 1 WHEN 'unverified' THEN 2 WHEN 'broken' THEN 3 ELSE 4 END, created_at DESC`, id); err != nil {
		return layoutdomain.Detail{}, fmt.Errorf("list layout links: %w", err)
	}

	matches, err := r.videoMatches(ctx, id)
	if err != nil {
		return layoutdomain.Detail{}, err
	}
	for _, match := range matches {
		switch match.MatchGroup {
		case "defense_replay":
			detail.DefenseReplays = append(detail.DefenseReplays, match)
		default:
			detail.AttackVideos = append(detail.AttackVideos, match)
		}
	}

	return detail, nil
}

func (r *LayoutRepository) CreateDraft(ctx context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error) {
	styleTags, err := encodeStringSlice(input.StyleTags)
	if err != nil {
		return layoutdomain.Detail{}, err
	}

	_, err = r.db.ExecContext(ctx, `
INSERT INTO base_layouts (
	id, title, th_level, layout_type, style_tags, source_type, source_url, review_status, quality_status, visibility
) VALUES (?, ?, ?, ?, CAST(? AS JSON), ?, NULLIF(?, ''), ?, ?, ?)`,
		input.ID,
		input.Title,
		input.THLevel,
		input.LayoutType,
		styleTags,
		input.SourceType,
		input.SourceURL,
		input.ReviewStatus,
		input.QualityStatus,
		input.Visibility,
	)
	if err != nil {
		return layoutdomain.Detail{}, fmt.Errorf("create layout draft: %w", err)
	}
	if err := r.createAuditLog(ctx, r.db, "layout", input.ID, "create_draft"); err != nil {
		return layoutdomain.Detail{}, err
	}

	return r.GetAdminDetail(ctx, input.ID)
}

func (r *LayoutRepository) AddLink(ctx context.Context, id string, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error) {
	if err := r.ensureLayoutExists(ctx, layoutID); err != nil {
		return layoutdomain.Link{}, err
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO layout_links (
	id, layout_id, link_type, url, link_status, source_type, source_url
) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''))`,
		id,
		layoutID,
		input.LinkType,
		input.URL,
		input.LinkStatus,
		input.SourceType,
		input.SourceURL,
	)
	if err != nil {
		return layoutdomain.Link{}, fmt.Errorf("add layout link: %w", err)
	}
	if err := r.createAuditLog(ctx, r.db, "layout_link", id, "add_link"); err != nil {
		return layoutdomain.Link{}, err
	}

	return r.getLink(ctx, id)
}

func (r *LayoutRepository) UpdateLink(ctx context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error) {
	result, err := r.db.ExecContext(ctx, `
UPDATE layout_links
SET link_status = ?, last_checked_at = CURRENT_TIMESTAMP(3), last_check_error = NULLIF(?, '')
WHERE id = ?`, input.LinkStatus, input.Note, linkID)
	if err != nil {
		return layoutdomain.Link{}, fmt.Errorf("update layout link: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return layoutdomain.Link{}, fmt.Errorf("read update affected rows: %w", err)
	}
	if affected == 0 {
		return layoutdomain.Link{}, layoutdomain.ErrLayoutNotFound
	}
	if err := r.createAuditLog(ctx, r.db, "layout_link", linkID, "update_link"); err != nil {
		return layoutdomain.Link{}, err
	}

	return r.getLink(ctx, linkID)
}

func (r *LayoutRepository) AddVideoMatch(ctx context.Context, matchID string, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error) {
	if err := r.ensureLayoutExists(ctx, layoutID); err != nil {
		return layoutdomain.VideoMatch{}, err
	}

	videoTags, err := encodeStringSlice(input.VideoTags)
	if err != nil {
		return layoutdomain.VideoMatch{}, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return layoutdomain.VideoMatch{}, fmt.Errorf("begin add video match: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	videoID, err := r.ensureVideo(ctx, tx, input)
	if err != nil {
		return layoutdomain.VideoMatch{}, err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO layout_video_matches (
	id, layout_id, video_id, timestamp_seconds, match_group, match_type, stars, destruction_percent, video_tags, confidence_score, review_status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CAST(? AS JSON), ?, ?)`,
		matchID,
		layoutID,
		videoID,
		input.TimestampSeconds,
		input.MatchGroup,
		input.MatchType,
		input.Stars,
		input.DestructionPercent,
		videoTags,
		input.ConfidenceScore,
		input.ReviewStatus,
	)
	if err != nil {
		return layoutdomain.VideoMatch{}, fmt.Errorf("insert video match: %w", err)
	}
	if err = r.createAuditLog(ctx, tx, "video_match", matchID, "add_video_match"); err != nil {
		return layoutdomain.VideoMatch{}, err
	}
	if err = tx.Commit(); err != nil {
		return layoutdomain.VideoMatch{}, fmt.Errorf("commit add video match: %w", err)
	}

	matches, err := r.videoMatchByID(ctx, matchID)
	if err != nil {
		return layoutdomain.VideoMatch{}, err
	}
	return matches, nil
}

func (r *LayoutRepository) ListReviewQueue(ctx context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error) {
	where, args := reviewQueueWhereClause(filter)
	baseQuery := `
FROM (
	SELECT 'layout' AS resource_type, id AS resource_id, title, review_status, quality_status, created_at, updated_at
	FROM base_layouts
	WHERE review_status <> 'reviewed' OR quality_status IN ('low_confidence', 'unknown')
	UNION ALL
	SELECT 'layout_link' AS resource_type, id AS resource_id, url AS title, link_status AS review_status, 'unknown' AS quality_status, created_at, COALESCE(last_checked_at, created_at) AS updated_at
	FROM layout_links
	WHERE link_status IN ('broken', 'unverified')
	UNION ALL
	SELECT 'image_search_job' AS resource_type, id AS resource_id, COALESCE(error_message, search_status) AS title, search_status AS review_status, COALESCE(screenshot_quality, 'unknown') AS quality_status, created_at, updated_at
	FROM image_search_jobs
	WHERE search_status IN ('low_confidence', 'failed')
) review_queue ` + where

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) "+baseQuery, args...); err != nil {
		return layoutdomain.ReviewQueueResult{}, fmt.Errorf("count review queue: %w", err)
	}

	var items []layoutdomain.ReviewQueueItem
	queryArgs := append(append([]any{}, args...), pagination.PageSize, pagination.Offset)
	if err := r.db.SelectContext(ctx, &items, `
SELECT resource_type, resource_id, title, review_status, quality_status, created_at, updated_at
`+baseQuery+`
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`, queryArgs...); err != nil {
		return layoutdomain.ReviewQueueResult{}, fmt.Errorf("list review queue: %w", err)
	}

	return layoutdomain.ReviewQueueResult{Items: items, Total: total}, nil
}

func (r *LayoutRepository) ListAuditLogs(ctx context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM admin_audit_logs"); err != nil {
		return layoutdomain.AuditLogResult{}, fmt.Errorf("count audit logs: %w", err)
	}

	var items []layoutdomain.AuditLog
	if err := r.db.SelectContext(ctx, &items, `
SELECT id, COALESCE(admin_id, '') AS admin_id, resource_type, resource_id, action, created_at
FROM admin_audit_logs
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, pagination.PageSize, pagination.Offset); err != nil {
		return layoutdomain.AuditLogResult{}, fmt.Errorf("list audit logs: %w", err)
	}
	return layoutdomain.AuditLogResult{Items: items, Total: total}, nil
}

func (r *LayoutRepository) UpdateReviewStatus(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	var (
		result layoutdomain.ReviewUpdateResult
		err    error
	)

	switch input.ResourceType {
	case "layout":
		result, err = r.updateLayoutReview(ctx, input)
	case "layout_image":
		result, err = r.updateLayoutImageReview(ctx, input)
	case "video":
		result, err = r.updateVideoReview(ctx, input)
	case "video_match", "layout_video_match":
		result, err = r.updateVideoMatchReview(ctx, input)
	case "image_search_job":
		result, err = r.updateImageSearchJobReview(ctx, input)
	case "layout_link":
		result, err = r.updateLayoutLinkReview(ctx, input)
	default:
		return layoutdomain.ReviewUpdateResult{}, layoutdomain.ErrReviewResourceNotFound
	}
	if err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}
	if err := r.createAuditLog(ctx, r.db, input.ResourceType, input.ResourceID, "update_review"); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}
	return result, nil
}

type contextExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *LayoutRepository) createAuditLog(ctx context.Context, execer contextExecer, resourceType string, resourceID string, action string) error {
	_, err := execer.ExecContext(ctx, `
INSERT INTO admin_audit_logs (id, resource_type, resource_id, action)
VALUES (?, ?, ?, ?)`, uuid.NewString(), resourceType, resourceID, action)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *LayoutRepository) updateLayoutReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE base_layouts
SET review_status = CASE WHEN ? <> '' THEN ? ELSE review_status END,
    quality_status = CASE WHEN ? <> '' THEN ? ELSE quality_status END,
    visibility = CASE WHEN ? <> '' THEN ? ELSE visibility END
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.QualityStatus, input.QualityStatus,
		input.Visibility, input.Visibility,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'layout' AS resource_type, id AS resource_id, review_status, quality_status, visibility, updated_at
FROM base_layouts
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated layout review")
	}
	return result, nil
}

func (r *LayoutRepository) updateLayoutImageReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE layout_images
SET review_status = CASE WHEN ? <> '' THEN ? ELSE review_status END,
    quality_status = CASE WHEN ? <> '' THEN ? ELSE quality_status END
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.QualityStatus, input.QualityStatus,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'layout_image' AS resource_type, id AS resource_id, review_status, quality_status, '' AS visibility, created_at AS updated_at
FROM layout_images
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated image review")
	}
	return result, nil
}

func (r *LayoutRepository) updateVideoReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE videos
SET review_status = CASE WHEN ? <> '' THEN ? ELSE review_status END,
    visibility = CASE WHEN ? <> '' THEN ? ELSE visibility END
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.Visibility, input.Visibility,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'video' AS resource_type, id AS resource_id, review_status, '' AS quality_status, visibility, created_at AS updated_at
FROM videos
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated video review")
	}
	return result, nil
}

func (r *LayoutRepository) updateVideoMatchReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE layout_video_matches
SET review_status = CASE WHEN ? <> '' THEN ? ELSE review_status END
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'video_match' AS resource_type, id AS resource_id, review_status, '' AS quality_status, '' AS visibility, created_at AS updated_at
FROM layout_video_matches
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated video match review")
	}
	return result, nil
}

func (r *LayoutRepository) updateImageSearchJobReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE image_search_jobs
SET search_status = CASE WHEN ? <> '' THEN ? ELSE search_status END,
    screenshot_quality = CASE WHEN ? <> '' THEN ? ELSE screenshot_quality END,
    updated_at = CURRENT_TIMESTAMP(3)
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.QualityStatus, input.QualityStatus,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'image_search_job' AS resource_type, id AS resource_id, search_status AS review_status, COALESCE(screenshot_quality, '') AS quality_status, '' AS visibility, updated_at
FROM image_search_jobs
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated image search job review")
	}
	return result, nil
}

func (r *LayoutRepository) updateLayoutLinkReview(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	if err := r.execReviewUpdate(ctx, `
UPDATE layout_links
SET link_status = CASE WHEN ? <> '' THEN ? ELSE link_status END,
    last_checked_at = CURRENT_TIMESTAMP(3),
    last_check_error = NULLIF(?, '')
WHERE id = ?`,
		input.ReviewStatus, input.ReviewStatus,
		input.Note,
		input.ResourceID,
	); err != nil {
		return layoutdomain.ReviewUpdateResult{}, err
	}

	var result layoutdomain.ReviewUpdateResult
	if err := r.db.GetContext(ctx, &result, `
SELECT 'layout_link' AS resource_type, id AS resource_id, link_status AS review_status, '' AS quality_status, '' AS visibility, COALESCE(last_checked_at, created_at) AS updated_at
FROM layout_links
WHERE id = ?`, input.ResourceID); err != nil {
		return layoutdomain.ReviewUpdateResult{}, reviewNotFound(err, "get updated layout link review")
	}
	return result, nil
}

func (r *LayoutRepository) execReviewUpdate(ctx context.Context, query string, args ...any) error {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update review status: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read review update affected rows: %w", err)
	}
	if affected == 0 {
		return layoutdomain.ErrReviewResourceNotFound
	}
	return nil
}

func reviewNotFound(err error, action string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return layoutdomain.ErrReviewResourceNotFound
	}
	return fmt.Errorf("%s: %w", action, err)
}

func (r *LayoutRepository) ensureVideo(ctx context.Context, tx *sqlx.Tx, input layoutdomain.VideoMatchInput) (string, error) {
	var existingID string
	if err := tx.GetContext(ctx, &existingID, "SELECT id FROM videos WHERE youtube_video_id = ?", input.YouTubeVideoID); err == nil {
		return existingID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("find video: %w", err)
	}

	videoID := uuid.NewString()
	_, err := tx.ExecContext(ctx, `
INSERT INTO videos (
	id, youtube_video_id, title, channel_name, source_type, source_url, review_status, visibility
) VALUES (?, ?, ?, NULLIF(?, ''), ?, NULLIF(?, ''), ?, 'public')`,
		videoID,
		input.YouTubeVideoID,
		input.VideoTitle,
		input.ChannelName,
		input.SourceType,
		input.SourceURL,
		input.ReviewStatus,
	)
	if err != nil {
		return "", fmt.Errorf("insert video: %w", err)
	}
	return videoID, nil
}

func (r *LayoutRepository) videoMatchByID(ctx context.Context, matchID string) (layoutdomain.VideoMatch, error) {
	var rows []layoutVideoMatchRow
	if err := r.db.SelectContext(ctx, &rows, `
SELECT
	lvm.id AS match_id,
	v.id AS video_id,
	v.youtube_video_id,
	v.title AS video_title,
	COALESCE(v.channel_name, '') AS channel_name,
	lvm.timestamp_seconds,
	lvm.match_group,
	lvm.match_type,
	lvm.stars,
	lvm.destruction_percent,
	COALESCE(CAST(lvm.video_tags AS CHAR), '[]') AS video_tags_json,
	lvm.confidence_score,
	lvm.review_status
FROM layout_video_matches lvm
JOIN videos v ON v.id = lvm.video_id
WHERE lvm.id = ?`, matchID); err != nil {
		return layoutdomain.VideoMatch{}, fmt.Errorf("get video match: %w", err)
	}
	if len(rows) == 0 {
		return layoutdomain.VideoMatch{}, layoutdomain.ErrLayoutNotFound
	}
	return rows[0].match()
}

func (r *LayoutRepository) GetAdminDetail(ctx context.Context, id string) (layoutdomain.Detail, error) {
	var row layoutDetailRow
	if err := r.db.GetContext(ctx, &row, `
SELECT
	id,
	title,
	th_level,
	layout_type,
	COALESCE(CAST(style_tags AS CHAR), '[]') AS style_tags_json,
	source_type,
	COALESCE(source_url, '') AS source_url,
	review_status,
	quality_status
FROM base_layouts
WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return layoutdomain.Detail{}, layoutdomain.ErrLayoutNotFound
		}
		return layoutdomain.Detail{}, fmt.Errorf("get admin layout: %w", err)
	}
	return row.detail()
}

func (r *LayoutRepository) ensureLayoutExists(ctx context.Context, id string) error {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM base_layouts WHERE id = ?)", id); err != nil {
		return fmt.Errorf("check layout exists: %w", err)
	}
	if !exists {
		return layoutdomain.ErrLayoutNotFound
	}
	return nil
}

func (r *LayoutRepository) getLink(ctx context.Context, id string) (layoutdomain.Link, error) {
	var link layoutdomain.Link
	if err := r.db.GetContext(ctx, &link, `
SELECT id, link_type, url, link_status, source_type, COALESCE(source_url, '') AS source_url, last_checked_at, COALESCE(last_check_error, '') AS last_check_error
FROM layout_links
WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return layoutdomain.Link{}, layoutdomain.ErrLayoutNotFound
		}
		return layoutdomain.Link{}, fmt.Errorf("get layout link: %w", err)
	}
	return link, nil
}

func (r *LayoutRepository) videoMatches(ctx context.Context, layoutID string) ([]layoutdomain.VideoMatch, error) {
	var rows []layoutVideoMatchRow
	if err := r.db.SelectContext(ctx, &rows, `
SELECT
	lvm.id AS match_id,
	v.id AS video_id,
	v.youtube_video_id,
	v.title AS video_title,
	COALESCE(v.channel_name, '') AS channel_name,
	lvm.timestamp_seconds,
	lvm.match_group,
	lvm.match_type,
	lvm.stars,
	lvm.destruction_percent,
	COALESCE(CAST(lvm.video_tags AS CHAR), '[]') AS video_tags_json,
	lvm.confidence_score,
	lvm.review_status
FROM layout_video_matches lvm
JOIN videos v ON v.id = lvm.video_id
WHERE lvm.layout_id = ? AND v.visibility = 'public'
ORDER BY CASE lvm.match_type WHEN 'exact' THEN 1 WHEN 'similar' THEN 2 ELSE 3 END, lvm.confidence_score DESC`, layoutID); err != nil {
		return nil, fmt.Errorf("list layout video matches: %w", err)
	}

	matches := make([]layoutdomain.VideoMatch, 0, len(rows))
	for _, row := range rows {
		match, err := row.match()
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, nil
}

// ListVideosByLayout returns public video matches for a layout, optionally filtered by match_group and match_type.
func (r *LayoutRepository) ListVideosByLayout(ctx context.Context, layoutID, matchGroup, matchType string) ([]layoutdomain.VideoMatch, error) {
	query := `
SELECT
	lvm.id AS match_id,
	v.id AS video_id,
	v.youtube_video_id,
	v.title AS video_title,
	COALESCE(v.channel_name, '') AS channel_name,
	lvm.timestamp_seconds,
	lvm.match_group,
	lvm.match_type,
	lvm.stars,
	lvm.destruction_percent,
	COALESCE(CAST(lvm.video_tags AS CHAR), '[]') AS video_tags_json,
	lvm.confidence_score,
	lvm.review_status
FROM layout_video_matches lvm
JOIN videos v ON v.id = lvm.video_id
WHERE lvm.layout_id = ? AND v.visibility = 'public'`
	args := []any{layoutID}
	if matchGroup != "" {
		query += " AND lvm.match_group = ?"
		args = append(args, matchGroup)
	}
	if matchType != "" {
		query += " AND lvm.match_type = ?"
		args = append(args, matchType)
	}
	query += " ORDER BY CASE lvm.match_type WHEN 'exact' THEN 1 WHEN 'similar' THEN 2 ELSE 3 END, lvm.confidence_score DESC"

	var rows []layoutVideoMatchRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list layout videos: %w", err)
	}
	matches := make([]layoutdomain.VideoMatch, 0, len(rows))
	for _, row := range rows {
		match, err := row.match()
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func layoutWhereClause(filter layoutdomain.ListFilter) (string, []any) {
	conditions := []string{"bl.visibility = 'public'"}
	args := make([]any, 0, 8)

	if filter.THLevel != nil {
		conditions = append(conditions, "bl.th_level = ?")
		args = append(args, *filter.THLevel)
	}
	if filter.LayoutType != "" {
		conditions = append(conditions, "bl.layout_type = ?")
		args = append(args, filter.LayoutType)
	}
	if filter.StyleTag != "" {
		conditions = append(conditions, "JSON_CONTAINS(bl.style_tags, JSON_QUOTE(?))")
		args = append(args, filter.StyleTag)
	}
	if filter.SourceType != "" {
		conditions = append(conditions, "bl.source_type = ?")
		args = append(args, filter.SourceType)
	}
	if filter.ReviewStatus != "" {
		conditions = append(conditions, "bl.review_status = ?")
		args = append(args, filter.ReviewStatus)
	}
	if filter.QualityStatus != "" {
		conditions = append(conditions, "bl.quality_status = ?")
		args = append(args, filter.QualityStatus)
	}
	if filter.LinkStatus != "" {
		if filter.LinkStatus == "missing" {
			conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM layout_links fl WHERE fl.layout_id = bl.id)")
		} else {
			conditions = append(conditions, "EXISTS (SELECT 1 FROM layout_links fl WHERE fl.layout_id = bl.id AND fl.link_status = ?)")
			args = append(args, filter.LinkStatus)
		}
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func reviewQueueWhereClause(filter layoutdomain.ReviewQueueFilter) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 3)
	if filter.ResourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	if filter.ReviewStatus != "" {
		conditions = append(conditions, "review_status = ?")
		args = append(args, filter.ReviewStatus)
	}
	if filter.QualityStatus != "" {
		conditions = append(conditions, "quality_status = ?")
		args = append(args, filter.QualityStatus)
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

type layoutCardRow struct {
	ID              string       `db:"id"`
	Title           string       `db:"title"`
	THLevel         int          `db:"th_level"`
	LayoutType      string       `db:"layout_type"`
	StyleTagsJSON   string       `db:"style_tags_json"`
	PrimaryImageURL string       `db:"primary_image_url"`
	SourceType      string       `db:"source_type"`
	ReviewStatus    string       `db:"review_status"`
	QualityStatus   string       `db:"quality_status"`
	LinkStatus      string       `db:"link_status"`
	UpdatedAt       sql.NullTime `db:"updated_at"`
}

func (r layoutCardRow) card() (layoutdomain.Card, error) {
	tags, err := decodeStringSlice(r.StyleTagsJSON)
	if err != nil {
		return layoutdomain.Card{}, err
	}
	return layoutdomain.Card{
		ID:              r.ID,
		Title:           r.Title,
		THLevel:         r.THLevel,
		LayoutType:      r.LayoutType,
		StyleTags:       tags,
		PrimaryImageURL: r.PrimaryImageURL,
		SourceType:      r.SourceType,
		ReviewStatus:    r.ReviewStatus,
		QualityStatus:   r.QualityStatus,
		LinkStatus:      r.LinkStatus,
		UpdatedAt:       nullTimeValue(r.UpdatedAt),
	}, nil
}

type layoutDetailRow struct {
	ID            string `db:"id"`
	Title         string `db:"title"`
	THLevel       int    `db:"th_level"`
	LayoutType    string `db:"layout_type"`
	StyleTagsJSON string `db:"style_tags_json"`
	SourceType    string `db:"source_type"`
	SourceURL     string `db:"source_url"`
	ReviewStatus  string `db:"review_status"`
	QualityStatus string `db:"quality_status"`
}

func (r layoutDetailRow) detail() (layoutdomain.Detail, error) {
	tags, err := decodeStringSlice(r.StyleTagsJSON)
	if err != nil {
		return layoutdomain.Detail{}, err
	}
	return layoutdomain.Detail{
		ID:            r.ID,
		Title:         r.Title,
		THLevel:       r.THLevel,
		LayoutType:    r.LayoutType,
		StyleTags:     tags,
		SourceType:    r.SourceType,
		SourceURL:     r.SourceURL,
		ReviewStatus:  r.ReviewStatus,
		QualityStatus: r.QualityStatus,
	}, nil
}

type layoutVideoMatchRow struct {
	MatchID            string          `db:"match_id"`
	VideoID            string          `db:"video_id"`
	YouTubeVideoID     string          `db:"youtube_video_id"`
	VideoTitle         string          `db:"video_title"`
	ChannelName        string          `db:"channel_name"`
	TimestampSeconds   int             `db:"timestamp_seconds"`
	MatchGroup         string          `db:"match_group"`
	MatchType          string          `db:"match_type"`
	Stars              sql.NullInt64   `db:"stars"`
	DestructionPercent sql.NullFloat64 `db:"destruction_percent"`
	VideoTagsJSON      string          `db:"video_tags_json"`
	ConfidenceScore    sql.NullFloat64 `db:"confidence_score"`
	ReviewStatus       string          `db:"review_status"`
}

func (r layoutVideoMatchRow) match() (layoutdomain.VideoMatch, error) {
	tags, err := decodeStringSlice(r.VideoTagsJSON)
	if err != nil {
		return layoutdomain.VideoMatch{}, err
	}
	match := layoutdomain.VideoMatch{
		MatchID:          r.MatchID,
		VideoID:          r.VideoID,
		YouTubeVideoID:   r.YouTubeVideoID,
		VideoTitle:       r.VideoTitle,
		ChannelName:      r.ChannelName,
		TimestampSeconds: r.TimestampSeconds,
		YouTubeURL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s&t=%ds", r.YouTubeVideoID, r.TimestampSeconds),
		MatchGroup:       r.MatchGroup,
		MatchType:        r.MatchType,
		VideoTags:        tags,
		ReviewStatus:     r.ReviewStatus,
	}
	if r.Stars.Valid {
		stars := int(r.Stars.Int64)
		match.Stars = &stars
	}
	if r.DestructionPercent.Valid {
		value := r.DestructionPercent.Float64
		match.DestructionPercent = &value
	}
	if r.ConfidenceScore.Valid {
		value := r.ConfidenceScore.Float64
		match.ConfidenceScore = &value
	}
	return match, nil
}

func decodeStringSlice(value string) ([]string, error) {
	if value == "" {
		return nil, nil
	}
	var result []string
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, fmt.Errorf("decode string slice: %w", err)
	}
	return result, nil
}

func encodeStringSlice(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encode string slice: %w", err)
	}
	return string(encoded), nil
}

func nullTimeValue(value sql.NullTime) (zero time.Time) {
	if value.Valid {
		return value.Time
	}
	return zero
}
