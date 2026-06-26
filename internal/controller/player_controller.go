package controller

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type PlayerReader interface {
	FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error)
	FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error)
}

type PlayerController struct {
	service PlayerReader
}

func NewPlayerController(service PlayerReader) *PlayerController {
	return &PlayerController{service: service}
}

func (c *PlayerController) GetPlayer(ctx *gin.Context) {
	playerTag := ctx.Param("tag")
	if playerTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	player, err := c.service.FetchPlayer(ctx.Request.Context(), playerTag)
	if err != nil {
		failPlayer(ctx, err)
		return
	}
	utils.OK(ctx, player)
}

func (c *PlayerController) GetBattleLog(ctx *gin.Context) {
	playerTag := ctx.Param("tag")
	if playerTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	log, err := c.service.FetchBattleLog(ctx.Request.Context(), playerTag)
	if err != nil {
		failPlayer(ctx, err)
		return
	}
	utils.OK(ctx, log)
}

func failPlayer(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, 422, int(utils.ErrInvalidField), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, 503, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorPlayerNotFound:
			utils.JSON(ctx, 404, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, 403, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, 502, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
