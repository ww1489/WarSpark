package layout

import (
	"errors"
	"time"
)

var ErrLayoutNotFound = errors.New("layout not found")
var ErrReviewResourceNotFound = errors.New("review resource not found")

// Canonical enum values from data-model.md.
var validReviewStatus = map[string]bool{
	"not_reviewed":   true,
	"reviewed":       true,
	"pending_review": true,
	"rejected":       true,
	"needs_update":   true,
}

var validQualityStatus = map[string]bool{
	"high_confidence":   true,
	"medium_confidence": true,
	"low_confidence":    true,
	"unknown":           true,
}

var validVisibility = map[string]bool{
	"public": true,
	"hidden": true,
}

var validLinkStatus = map[string]bool{
	"active":     true,
	"broken":     true,
	"unverified": true,
	"missing":    true,
}

var validLinkType = map[string]bool{
	"official_open_layout": true,
	"source_page":          true,
	"backup":               true,
}

func ValidReviewStatus(v string) bool  { return validReviewStatus[v] }
func ValidQualityStatus(v string) bool { return validQualityStatus[v] }
func ValidVisibility(v string) bool    { return validVisibility[v] }
func ValidLinkStatus(v string) bool    { return validLinkStatus[v] }
func ValidLinkType(v string) bool      { return validLinkType[v] }

type ListFilter struct {
	THLevel       *int
	LayoutType    string
	StyleTag      string
	SourceType    string
	ReviewStatus  string
	QualityStatus string
	LinkStatus    string
}

type ListResult struct {
	Items []Card
	Total int
}

type CreateInput struct {
	ID            string   `json:"-"`
	Title         string   `json:"title"`
	THLevel       int      `json:"th_level"`
	LayoutType    string   `json:"layout_type"`
	StyleTags     []string `json:"style_tags"`
	SourceType    string   `json:"source_type"`
	SourceURL     string   `json:"source_url"`
	ReviewStatus  string   `json:"review_status"`
	QualityStatus string   `json:"quality_status"`
	Visibility    string   `json:"visibility"`
}

type LinkInput struct {
	LinkType   string `json:"link_type"`
	URL        string `json:"url"`
	LinkStatus string `json:"link_status"`
	SourceType string `json:"source_type"`
	SourceURL  string `json:"source_url"`
}

type LinkUpdateInput struct {
	LinkStatus string `json:"link_status"`
	Note       string `json:"note"`
}

type ReviewUpdateInput struct {
	ResourceType  string `json:"-"`
	ResourceID    string `json:"-"`
	ReviewStatus  string `json:"review_status"`
	QualityStatus string `json:"quality_status"`
	Visibility    string `json:"visibility"`
	Note          string `json:"note"`
}

type ReviewUpdateResult struct {
	ResourceType  string    `json:"resource_type" db:"resource_type"`
	ResourceID    string    `json:"resource_id" db:"resource_id"`
	ReviewStatus  string    `json:"review_status,omitempty" db:"review_status"`
	QualityStatus string    `json:"quality_status,omitempty" db:"quality_status"`
	Visibility    string    `json:"visibility,omitempty" db:"visibility"`
	UpdatedAt     time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type VideoMatchInput struct {
	YouTubeVideoID     string   `json:"youtube_video_id"`
	VideoTitle         string   `json:"video_title"`
	ChannelName        string   `json:"channel_name"`
	TimestampSeconds   int      `json:"timestamp_seconds"`
	MatchGroup         string   `json:"match_group"`
	MatchType          string   `json:"match_type"`
	Stars              *int     `json:"stars"`
	DestructionPercent *float64 `json:"destruction_percent"`
	VideoTags          []string `json:"video_tags"`
	ConfidenceScore    *float64 `json:"confidence_score"`
	ReviewStatus       string   `json:"review_status"`
	SourceType         string   `json:"source_type"`
	SourceURL          string   `json:"source_url"`
}

type ReviewQueueFilter struct {
	ResourceType  string
	ReviewStatus  string
	QualityStatus string
}

type ReviewQueueResult struct {
	Items []ReviewQueueItem
	Total int
}

type ReviewQueueItem struct {
	ResourceType  string    `json:"resource_type" db:"resource_type"`
	ResourceID    string    `json:"resource_id" db:"resource_id"`
	Title         string    `json:"title" db:"title"`
	ReviewStatus  string    `json:"review_status" db:"review_status"`
	QualityStatus string    `json:"quality_status" db:"quality_status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type AuditLogResult struct {
	Items []AuditLog
	Total int
}

type AuditLog struct {
	ID           string    `json:"audit_id" db:"id"`
	AdminID      string    `json:"admin_id,omitempty" db:"admin_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Action       string    `json:"action" db:"action"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Card struct {
	ID              string    `json:"layout_id" db:"id"`
	Title           string    `json:"title" db:"title"`
	THLevel         int       `json:"th_level" db:"th_level"`
	LayoutType      string    `json:"layout_type" db:"layout_type"`
	StyleTags       []string  `json:"style_tags" db:"-"`
	PrimaryImageURL string    `json:"primary_image_url" db:"primary_image_url"`
	SourceType      string    `json:"source_type" db:"source_type"`
	ReviewStatus    string    `json:"review_status" db:"review_status"`
	QualityStatus   string    `json:"quality_status" db:"quality_status"`
	LinkStatus      string    `json:"link_status" db:"link_status"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Detail struct {
	ID             string       `json:"layout_id" db:"id"`
	Title          string       `json:"title" db:"title"`
	THLevel        int          `json:"th_level" db:"th_level"`
	LayoutType     string       `json:"layout_type" db:"layout_type"`
	StyleTags      []string     `json:"style_tags" db:"-"`
	SourceType     string       `json:"source_type" db:"source_type"`
	SourceURL      string       `json:"source_url,omitempty" db:"source_url"`
	ReviewStatus   string       `json:"review_status" db:"review_status"`
	QualityStatus  string       `json:"quality_status" db:"quality_status"`
	Images         []Image      `json:"images"`
	Links          []Link       `json:"links"`
	AttackVideos   []VideoMatch `json:"attack_videos"`
	DefenseReplays []VideoMatch `json:"defense_replays"`
	SimilarLayouts []Card       `json:"similar_layouts"`
}

type Image struct {
	ID            string `json:"image_id" db:"id"`
	ImageURL      string `json:"image_url" db:"image_url"`
	SourceType    string `json:"source_type,omitempty" db:"source_type"`
	SourceURL     string `json:"source_url,omitempty" db:"source_url"`
	Width         *int   `json:"width,omitempty" db:"width"`
	Height        *int   `json:"height,omitempty" db:"height"`
	ImageRole     string `json:"image_role" db:"image_role"`
	ReviewStatus  string `json:"review_status,omitempty" db:"review_status"`
	QualityStatus string `json:"quality_status,omitempty" db:"quality_status"`
}

type Link struct {
	ID             string     `json:"link_id" db:"id"`
	LinkType       string     `json:"link_type" db:"link_type"`
	URL            string     `json:"url" db:"url"`
	LinkStatus     string     `json:"link_status" db:"link_status"`
	SourceType     string     `json:"source_type,omitempty" db:"source_type"`
	SourceURL      string     `json:"source_url,omitempty" db:"source_url"`
	LastCheckedAt  *time.Time `json:"last_checked_at,omitempty" db:"last_checked_at"`
	LastCheckError string     `json:"last_check_error,omitempty" db:"last_check_error"`
}

type VideoMatch struct {
	MatchID            string   `json:"match_id" db:"match_id"`
	VideoID            string   `json:"video_id" db:"video_id"`
	YouTubeVideoID     string   `json:"youtube_video_id" db:"youtube_video_id"`
	VideoTitle         string   `json:"video_title" db:"video_title"`
	ChannelName        string   `json:"channel_name,omitempty" db:"channel_name"`
	TimestampSeconds   int      `json:"timestamp_seconds" db:"timestamp_seconds"`
	YouTubeURL         string   `json:"youtube_url" db:"-"`
	MatchGroup         string   `json:"match_group" db:"match_group"`
	MatchType          string   `json:"match_type" db:"match_type"`
	Stars              *int     `json:"stars,omitempty" db:"stars"`
	DestructionPercent *float64 `json:"destruction_percent,omitempty" db:"destruction_percent"`
	VideoTags          []string `json:"video_tags" db:"-"`
	ConfidenceScore    *float64 `json:"confidence_score,omitempty" db:"confidence_score"`
	ReviewStatus       string   `json:"review_status" db:"review_status"`
}
