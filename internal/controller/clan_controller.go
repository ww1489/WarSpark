package controller

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type ClanReader interface {
	FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error)
}

type ClanController struct {
	service ClanReader
}

func NewClanController(service ClanReader) *ClanController {
	return &ClanController{service: service}
}

// GetClan 获取部落详细信息。
//
// @Summary 获取部落信息
// @Tags clan
// @Produce json
// @Param tag path string true "部落标签"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 422 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/clans/{tag} [get]
func (c *ClanController) GetClan(ctx *gin.Context) {
	clanTag := ctx.Param("tag")
	if clanTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "tag is required"))
		return
	}

	detail, err := c.service.FetchClan(ctx.Request.Context(), clanTag)
	if err != nil {
		failClan(ctx, err)
		return
	}
	utils.OK(ctx, detail)
}

func failClan(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, 422, int(utils.ErrInvalidField), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, 503, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorClanNotFound:
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
