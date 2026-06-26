package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type RankingReader interface {
	GetLocations(ctx context.Context) (ranking.LocationListResponse, error)
	GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error)
	GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error)
	GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error)
	GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error)
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error)
}

type RankingController struct {
	service RankingReader
}

func NewRankingController(service RankingReader) *RankingController {
	return &RankingController{service: service}
}

// GetLocations 获取位置列表。
//
// @Summary 获取位置列表
// @Tags ranking
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations [get]
func (c *RankingController) GetLocations(ctx *gin.Context) {
	resp, err := c.service.GetLocations(ctx.Request.Context())
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetClanRanking 获取指定位置的部落排名。
//
// @Summary 获取部落排名
// @Tags ranking
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id}/rankings/clans [get]
func (c *RankingController) GetClanRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanRanking(ctx.Request.Context(), locationID)
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetPlayerRanking 获取指定位置的玩家排名。
//
// @Summary 获取玩家排名
// @Tags ranking
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id}/rankings/players [get]
func (c *RankingController) GetPlayerRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetPlayerRanking(ctx.Request.Context(), locationID)
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetClanCapitalRanking 获取指定位置的部落都城排名。
//
// @Summary 获取部落都城排名
// @Tags ranking
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id}/rankings/clans-capital [get]
func (c *RankingController) GetClanCapitalRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanCapitalRanking(ctx.Request.Context(), locationID)
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetClanBuilderBaseRanking 获取指定位置的部落夜世界排名。
//
// @Summary 获取部落夜世界排名
// @Tags ranking
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id}/rankings/clans-builder-base [get]
func (c *RankingController) GetClanBuilderBaseRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetClanBuilderBaseRanking(ctx.Request.Context(), locationID)
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetPlayerBuilderBaseRanking 获取指定位置的玩家夜世界排名。
//
// @Summary 获取玩家夜世界排名
// @Tags ranking
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id}/rankings/players-builder-base [get]
func (c *RankingController) GetPlayerBuilderBaseRanking(ctx *gin.Context) {
	locationID := ctx.Param("id")
	resp, err := c.service.GetPlayerBuilderBaseRanking(ctx.Request.Context(), locationID)
	if err != nil {
		failRanking(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func failRanking(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorLocationNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
