package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
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

func (c *LeagueController) GetLeagues(ctx *gin.Context) {
	resp, err := c.service.GetLeagues(ctx.Request.Context())
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func (c *LeagueController) GetLeague(ctx *gin.Context) {
	id := ctx.Param("id")
	l, err := c.service.GetLeague(ctx.Request.Context(), id)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, l)
}

func (c *LeagueController) GetLeagueSeasons(ctx *gin.Context) {
	id := ctx.Param("id")
	resp, err := c.service.GetLeagueSeasons(ctx.Request.Context(), id)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

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

func (c *LeagueController) GetLeagueTier(ctx *gin.Context) {
	tierID := ctx.Param("tier")
	t, err := c.service.GetLeagueTier(ctx.Request.Context(), tierID)
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, t)
}

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

func (c *LeagueController) GetWarLeagues(ctx *gin.Context) {
	resp, err := c.service.GetWarLeagues(ctx.Request.Context())
	if err != nil {
		failLeague(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

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
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorLeagueNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
