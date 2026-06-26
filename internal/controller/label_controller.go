package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/label"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LabelReader interface {
	GetClanLabels(ctx context.Context) (label.LabelListResponse, error)
	GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error)
}

type LabelController struct {
	service LabelReader
}

func NewLabelController(service LabelReader) *LabelController {
	return &LabelController{service: service}
}

func (c *LabelController) GetClanLabels(ctx *gin.Context) {
	resp, err := c.service.GetClanLabels(ctx.Request.Context())
	if err != nil {
		failLabel(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func (c *LabelController) GetPlayerLabels(ctx *gin.Context) {
	resp, err := c.service.GetPlayerLabels(ctx.Request.Context())
	if err != nil {
		failLabel(ctx, err)
		return
	}
	utils.OK(ctx, resp)
}

func failLabel(ctx *gin.Context, err error) {
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
