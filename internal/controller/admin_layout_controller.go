package controller

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

type AdminLayoutWriter interface {
	CreateLayoutDraft(ctx context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error)
	AddLayoutLink(ctx context.Context, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error)
	UpdateLayoutLink(ctx context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error)
	AddLayoutVideoMatch(ctx context.Context, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error)
	ListReviewQueue(ctx context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error)
	ListAuditLogs(ctx context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error)
	UpdateReviewStatus(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error)
}

type AdminLayoutController struct {
	service AdminLayoutWriter
}

func NewAdminLayoutController(service AdminLayoutWriter) *AdminLayoutController {
	return &AdminLayoutController{service: service}
}

// CreateDraft creates an internal layout draft.
//
// @Summary Create layout draft
// @Tags admin-layouts
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/layouts [post]
func (c *AdminLayoutController) CreateDraft(ctx *gin.Context) {
	var input layoutdomain.CreateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid request body"))
		return
	}
	normalizeCreateInput(&input)
	if input.Title == "" || input.THLevel <= 0 || input.LayoutType == "" || input.SourceType == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "title, th_level, layout_type, and source_type are required"))
		return
	}

	detail, err := c.service.CreateLayoutDraft(ctx.Request.Context(), input)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, detail)
}

// AddLink adds a layout link in the internal admin flow.
//
// @Summary Add layout link
// @Tags admin-layouts
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/layouts/{layout_id}/links [post]
func (c *AdminLayoutController) AddLink(ctx *gin.Context) {
	layoutID := ctx.Param("layout_id")
	if layoutID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "layout id is required"))
		return
	}

	var input layoutdomain.LinkInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid request body"))
		return
	}
	normalizeLinkInput(&input)
	if input.LinkType == "" || input.URL == "" || input.SourceType == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "link_type, url, and source_type are required"))
		return
	}

	link, err := c.service.AddLayoutLink(ctx.Request.Context(), layoutID, input)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, link)
}

// UpdateLink updates link status and audit note.
//
// @Summary Update layout link
// @Tags admin-layouts
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/layout-links/{link_id} [patch]
func (c *AdminLayoutController) UpdateLink(ctx *gin.Context) {
	linkID := ctx.Param("link_id")
	if linkID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "link id is required"))
		return
	}

	var input layoutdomain.LinkUpdateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid request body"))
		return
	}
	if input.LinkStatus == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "link_status is required"))
		return
	}
	if !layoutdomain.ValidLinkStatus(input.LinkStatus) {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid link_status"))
		return
	}

	link, err := c.service.UpdateLayoutLink(ctx.Request.Context(), linkID, input)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, link)
}

// UpdateReviewStatus updates generic review, quality, and visibility state.
//
// @Summary Update review status
// @Tags admin-layouts
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/review/{resource_type}/{resource_id} [patch]
func (c *AdminLayoutController) UpdateReviewStatus(ctx *gin.Context) {
	resourceType := ctx.Param("resource_type")
	resourceID := ctx.Param("resource_id")
	if resourceType == "" || resourceID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "resource type and resource id are required"))
		return
	}
	if !supportedReviewResourceType(resourceType) {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "unsupported resource_type"))
		return
	}

	var input layoutdomain.ReviewUpdateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid request body"))
		return
	}
	input.ResourceType = resourceType
	input.ResourceID = resourceID
	if input.ReviewStatus == "" && input.QualityStatus == "" && input.Visibility == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "review_status, quality_status, or visibility is required"))
		return
	}
	if input.ReviewStatus != "" && !layoutdomain.ValidReviewStatus(input.ReviewStatus) {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid review_status"))
		return
	}
	if input.QualityStatus != "" && !layoutdomain.ValidQualityStatus(input.QualityStatus) {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid quality_status"))
		return
	}
	if input.Visibility != "" && !layoutdomain.ValidVisibility(input.Visibility) {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid visibility"))
		return
	}

	result, err := c.service.UpdateReviewStatus(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, layoutdomain.ErrReviewResourceNotFound) {
			utils.Fail(ctx, utils.NewError(utils.ErrNotFound, "review resource not found"))
			return
		}
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, result)
}

// AddVideoMatch adds a YouTube timestamp association for a layout.
//
// @Summary Add layout video match
// @Tags admin-layouts
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/layouts/{layout_id}/video-matches [post]
func (c *AdminLayoutController) AddVideoMatch(ctx *gin.Context) {
	layoutID := ctx.Param("layout_id")
	if layoutID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "layout id is required"))
		return
	}

	var input layoutdomain.VideoMatchInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "invalid request body"))
		return
	}
	normalizeVideoMatchInput(&input)
	if input.YouTubeVideoID == "" || input.VideoTitle == "" || input.TimestampSeconds < 0 || input.MatchGroup == "" || input.MatchType == "" || input.SourceType == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "youtube_video_id, video_title, timestamp_seconds, match_group, match_type, and source_type are required"))
		return
	}

	match, err := c.service.AddLayoutVideoMatch(ctx.Request.Context(), layoutID, input)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, match)
}

// ListReviewQueue returns internal items that need review or repair.
//
// @Summary List review queue
// @Tags admin-layouts
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/review-queue [get]
func (c *AdminLayoutController) ListReviewQueue(ctx *gin.Context) {
	filter := layoutdomain.ReviewQueueFilter{
		ResourceType:  ctx.Query("resource_type"),
		ReviewStatus:  ctx.Query("review_status"),
		QualityStatus: ctx.Query("quality_status"),
	}
	pagination := utils.ParsePagination(ctx)
	result, err := c.service.ListReviewQueue(ctx.Request.Context(), filter, pagination)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, paginated(result.Items, pagination, result.Total))
}

// ListAuditLogs returns internal admin write logs.
//
// @Summary List admin audit logs
// @Tags admin-layouts
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/admin/audit-logs [get]
func (c *AdminLayoutController) ListAuditLogs(ctx *gin.Context) {
	pagination := utils.ParsePagination(ctx)
	result, err := c.service.ListAuditLogs(ctx.Request.Context(), pagination)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, paginated(result.Items, pagination, result.Total))
}

func normalizeCreateInput(input *layoutdomain.CreateInput) {
	if input.ReviewStatus == "" {
		input.ReviewStatus = "pending_review"
	}
	if input.QualityStatus == "" {
		input.QualityStatus = "unknown"
	}
	if input.Visibility == "" {
		input.Visibility = "hidden"
	}
	if !layoutdomain.ValidReviewStatus(input.ReviewStatus) {
		input.ReviewStatus = "pending_review"
	}
	if !layoutdomain.ValidQualityStatus(input.QualityStatus) {
		input.QualityStatus = "unknown"
	}
	if !layoutdomain.ValidVisibility(input.Visibility) {
		input.Visibility = "hidden"
	}
}

func normalizeLinkInput(input *layoutdomain.LinkInput) {
	if input.LinkStatus == "" {
		input.LinkStatus = "unverified"
	}
	if !layoutdomain.ValidLinkStatus(input.LinkStatus) {
		input.LinkStatus = "unverified"
	}
}

func normalizeVideoMatchInput(input *layoutdomain.VideoMatchInput) {
	if input.ReviewStatus == "" {
		input.ReviewStatus = "pending_review"
	}
}

func paginated(items any, pagination utils.Pagination, total int) gin.H {
	return gin.H{
		"items": items,
		"pagination": gin.H{
			"page":      pagination.Page,
			"page_size": pagination.PageSize,
			"total":     total,
		},
	}
}

func supportedReviewResourceType(resourceType string) bool {
	switch resourceType {
	case "layout", "layout_image", "video", "video_match", "layout_video_match", "image_search_job", "layout_link":
		return true
	default:
		return false
	}
}
