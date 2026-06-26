package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/capital"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/utils"
)

type CapitalReader interface {
	GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, error)
	GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, error)
	GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, error)
	GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, error)
	GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, error)
}

type CapitalController struct {
	service CapitalReader
}

func NewCapitalController(service CapitalReader) *CapitalController {
	return &CapitalController{service: service}
}

// GetCapitalRaidSeasons 获取部落都城突袭赛季数据。
//
// @Summary 获取都城突袭赛季
// @Tags capital
// @Produce json
// @Param tag path string true "部落标签"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/clans/{tag}/capital-raid-seasons [get]
func (c *CapitalController) GetCapitalRaidSeasons(ctx *gin.Context) {
	clanTag := ctx.Param("tag")
	resp, err := c.service.GetCapitalRaidSeasons(ctx.Request.Context(), clanTag)
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetCapitalLeagues 获取都城联赛列表。
//
// @Summary 获取都城联赛列表
// @Tags capital
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/capital-leagues [get]
func (c *CapitalController) GetCapitalLeagues(ctx *gin.Context) {
	resp, err := c.service.GetCapitalLeagues(ctx.Request.Context())
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetCapitalLeague 获取都城联赛详情。
//
// @Summary 获取都城联赛详情
// @Tags capital
// @Produce json
// @Param id path string true "联赛ID"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/capital-leagues/{id} [get]
func (c *CapitalController) GetCapitalLeague(ctx *gin.Context) {
	leagueID := ctx.Param("id")
	resp, err := c.service.GetCapitalLeague(ctx.Request.Context(), leagueID)
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetBuilderBaseLeagues 获取夜世界联赛列表。
//
// @Summary 获取夜世界联赛列表
// @Tags capital
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/builder-base-leagues [get]
func (c *CapitalController) GetBuilderBaseLeagues(ctx *gin.Context) {
	resp, err := c.service.GetBuilderBaseLeagues(ctx.Request.Context())
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetBuilderBaseLeague 获取夜世界联赛详情。
//
// @Summary 获取夜世界联赛详情
// @Tags capital
// @Produce json
// @Param id path string true "联赛ID"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/builder-base-leagues/{id} [get]
func (c *CapitalController) GetBuilderBaseLeague(ctx *gin.Context) {
	leagueID := ctx.Param("id")
	resp, err := c.service.GetBuilderBaseLeague(ctx.Request.Context(), leagueID)
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func failCapital(ctx *gin.Context, err error) {
	var warErr dmerrors.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case dmerrors.ErrCodeAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case dmerrors.ErrCodeAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
