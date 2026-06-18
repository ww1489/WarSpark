package controller

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LayoutReader interface {
	ListLayouts(ctx context.Context, filter layoutdomain.ListFilter, pagination utils.Pagination) (layoutdomain.ListResult, error)
	GetLayout(ctx context.Context, id string) (layoutdomain.Detail, error)
	ListVideos(ctx context.Context, layoutID, matchGroup, matchType string) ([]layoutdomain.VideoMatch, error)
}

type LayoutController struct {
	service LayoutReader
}

func NewLayoutController(service LayoutReader) *LayoutController {
	return &LayoutController{service: service}
}

// List returns public layout cards with MVP filters.
//
// @Summary List base layouts
// @Tags layouts
// @Produce json
// @Param th_level query int false "Town Hall level"
// @Param layout_type query string false "Layout type"
// @Param style_tag query string false "Style tag"
// @Param source_type query string false "Source type"
// @Param review_status query string false "Review status"
// @Param quality_status query string false "Quality status"
// @Param link_status query string false "Link status"
// @Success 200 {object} utils.Response
// @Failure 422 {object} utils.Response
// @Router /api/v1/layouts [get]
func (c *LayoutController) List(ctx *gin.Context) {
	filter, err := layoutFilter(ctx)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	pagination := utils.ParsePagination(ctx)

	result, err := c.service.ListLayouts(ctx.Request.Context(), filter, pagination)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}

	utils.OK(ctx, gin.H{
		"items": result.Items,
		"pagination": gin.H{
			"page":      pagination.Page,
			"page_size": pagination.PageSize,
			"total":     result.Total,
		},
	})
}

// Get returns one layout detail.
//
// @Summary Get base layout
// @Tags layouts
// @Produce json
// @Param layout_id path string true "Layout ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/layouts/{layout_id} [get]
func (c *LayoutController) Get(ctx *gin.Context) {
	id := ctx.Param("layout_id")
	if id == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "layout id is required"))
		return
	}

	detail, err := c.service.GetLayout(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, layoutdomain.ErrLayoutNotFound) {
			utils.Fail(ctx, utils.NewError(utils.ErrNotFound, "layout not found"))
			return
		}
		utils.Fail(ctx, err)
		return
	}

	utils.OK(ctx, detail)
}


// ListVideos returns public video matches for a layout, optionally filtered by match_group and match_type.
//
// @Summary List layout videos
// @Tags layouts
// @Produce json
// @Param layout_id path string true "Layout ID"
// @Param match_group query string false "attack_video or defense_replay"
// @Param match_type query string false "exact, similar, or same_th"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/layouts/{layout_id}/videos [get]
func (c *LayoutController) ListVideos(ctx *gin.Context) {
	layoutID := ctx.Param("layout_id")
	if layoutID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "layout id is required"))
		return
	}

	matchGroup := ctx.Query("match_group")
	matchType := ctx.Query("match_type")
	videos, err := c.service.ListVideos(ctx.Request.Context(), layoutID, matchGroup, matchType)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}
	utils.OK(ctx, gin.H{"items": videos})
}
func layoutFilter(ctx *gin.Context) (layoutdomain.ListFilter, error) {
	filter := layoutdomain.ListFilter{
		LayoutType:    ctx.Query("layout_type"),
		StyleTag:      ctx.Query("style_tag"),
		SourceType:    ctx.Query("source_type"),
		ReviewStatus:  ctx.Query("review_status"),
		QualityStatus: ctx.Query("quality_status"),
		LinkStatus:    ctx.Query("link_status"),
	}
	if value := ctx.Query("th_level"); value != "" {
		thLevel, err := strconv.Atoi(value)
		if err != nil || thLevel <= 0 {
			return layoutdomain.ListFilter{}, utils.NewError(utils.ErrInvalidField, "invalid th_level")
		}
		filter.THLevel = &thLevel
	}
	return filter, nil
}
