package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/capital"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
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

func (c *CapitalController) GetCapitalRaidSeasons(ctx *gin.Context) {
	clanTag := ctx.Param("tag")
	resp, err := c.service.GetCapitalRaidSeasons(ctx.Request.Context(), clanTag)
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func (c *CapitalController) GetCapitalLeagues(ctx *gin.Context) {
	resp, err := c.service.GetCapitalLeagues(ctx.Request.Context())
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func (c *CapitalController) GetCapitalLeague(ctx *gin.Context) {
	leagueID := ctx.Param("id")
	resp, err := c.service.GetCapitalLeague(ctx.Request.Context(), leagueID)
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func (c *CapitalController) GetBuilderBaseLeagues(ctx *gin.Context) {
	resp, err := c.service.GetBuilderBaseLeagues(ctx.Request.Context())
	if err != nil {
		failCapital(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

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
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
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
