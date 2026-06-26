package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/domain/league"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LeagueReader interface {
	GetLeagues(ctx context.Context) (league.LeagueListResponse, error)
	GetLeague(ctx context.Context, id string) (league.League, error)
	GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error)
	GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error)
	GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error)
	GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error)
	GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error)
	GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error)
	GetWarLeague(ctx context.Context, id string) (league.WarLeague, error)
}

type LeagueController struct {
	service LeagueReader
}

func NewLeagueController(service LeagueReader) *LeagueController {
	return &LeagueController{service: service}
}

// GetLeagues 获取联赛列表。
//
// @Summary 获取联赛列表
// @Tags league
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues [get]
func (c *LeagueController) GetLeagues(ctx *gin.Context) {
	resp, err := c.service.GetLeagues(ctx.Request.Context())
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetLeague 获取联赛详情。
//
// @Summary 获取联赛详情
// @Tags league
// @Produce json
// @Param id path string true "联赛ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id} [get]
func (c *LeagueController) GetLeague(ctx *gin.Context) {
	id := ctx.Param("id")
	l, err := c.service.GetLeague(ctx.Request.Context(), id)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, l)
}

// GetLeagueSeasons 获取联赛赛季列表。
//
// @Summary 获取联赛赛季列表
// @Tags league
// @Produce json
// @Param id path string true "联赛ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id}/seasons [get]
func (c *LeagueController) GetLeagueSeasons(ctx *gin.Context) {
	id := ctx.Param("id")
	resp, err := c.service.GetLeagueSeasons(ctx.Request.Context(), id)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetLeagueSeasonRankings 获取联赛赛季排名。
//
// @Summary 获取联赛赛季排名
// @Tags league
// @Produce json
// @Param id path string true "联赛ID"
// @Param season path string true "赛季ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id}/seasons/{season}/rankings [get]
func (c *LeagueController) GetLeagueSeasonRankings(ctx *gin.Context) {
	id := ctx.Param("id")
	season := ctx.Param("season")
	resp, err := c.service.GetLeagueSeasonRankings(ctx.Request.Context(), id, season)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetLeagueTiers 获取联赛赛季段位列表。
//
// @Summary 获取联赛赛季段位列表
// @Tags league
// @Produce json
// @Param id path string true "联赛ID"
// @Param season path string true "赛季ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id}/seasons/{season}/tiers [get]
func (c *LeagueController) GetLeagueTiers(ctx *gin.Context) {
	id := ctx.Param("id")
	season := ctx.Param("season")
	resp, err := c.service.GetLeagueTiers(ctx.Request.Context(), id, season)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetLeagueTier 获取联赛赛季段位详情。
//
// @Summary 获取联赛赛季段位详情
// @Tags league
// @Produce json
// @Param tier path string true "段位ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id}/seasons/{season}/tiers/{tier} [get]
func (c *LeagueController) GetLeagueTier(ctx *gin.Context) {
	tierID := ctx.Param("tier")
	t, err := c.service.GetLeagueTier(ctx.Request.Context(), tierID)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, t)
}

// GetLeagueHistory 获取玩家联赛赛季历史。
//
// @Summary 获取玩家联赛赛季历史
// @Tags league
// @Produce json
// @Param player_tag query string true "玩家标签"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/leagues/{id}/seasons/{season}/tiers/{tier}/history [get]
func (c *LeagueController) GetLeagueHistory(ctx *gin.Context) {
	playerTag := ctx.Query("player_tag")
	if playerTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "player_tag is required"))
		return
	}
	resp, err := c.service.GetLeagueHistory(ctx.Request.Context(), playerTag)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetWarLeagues 获取战争联赛列表。
//
// @Summary 获取战争联赛列表
// @Tags league
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/war-leagues [get]
func (c *LeagueController) GetWarLeagues(ctx *gin.Context) {
	resp, err := c.service.GetWarLeagues(ctx.Request.Context())
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

// GetWarLeague 获取战争联赛详情。
//
// @Summary 获取战争联赛详情
// @Tags league
// @Produce json
// @Param id path string true "联赛ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/war-leagues/{id} [get]
func (c *LeagueController) GetWarLeague(ctx *gin.Context) {
	id := ctx.Param("id")
	l, err := c.service.GetWarLeague(ctx.Request.Context(), id)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, l)
}

func failLeague(ctx *gin.Context, err error) {
	var warErr dmerrors.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case dmerrors.ErrCodeLeagueNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case dmerrors.ErrCodeAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case dmerrors.ErrCodeAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		case dmerrors.ErrCodeInvalidTag:
			utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
