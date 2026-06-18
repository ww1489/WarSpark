package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
)

type WarReader interface {
	FetchCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, error)
	ListMembers(ctx context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error)
}

type WarController struct {
	service WarReader
}

func NewWarController(service WarReader) *WarController {
	return &WarController{service: service}
}

// GetCurrent returns current war data from the official Clash of Clans API.
//
// @Summary Get current clan war
// @Tags war
// @Produce json
// @Param clan_tag query string true "Clan tag"
// @Success 200 {object} utils.Response
// @Failure 422 {object} utils.Response
// @Router /api/v1/war/current [get]
func (c *WarController) GetCurrent(ctx *gin.Context) {
	clanTag := ctx.Query("clan_tag")
	if clanTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "clan_tag is required"))
		return
	}

	snapshot, err := c.service.FetchCurrentWar(ctx.Request.Context(), clanTag)
	if err != nil {
		failWar(ctx, err)
		return
	}
	utils.OK(ctx, snapshot)
}

// ListMembers returns members for a saved war snapshot.
//
// @Summary List war snapshot members
// @Tags war
// @Produce json
// @Param war_snapshot_id path string true "War snapshot ID"
// @Param side query string false "clan or opponent"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/war/snapshots/{war_snapshot_id}/members [get]
func (c *WarController) ListMembers(ctx *gin.Context) {
	snapshotID := ctx.Param("war_snapshot_id")
	if snapshotID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "war_snapshot_id is required"))
		return
	}
	side := ctx.Query("side")
	if side != "" && side != "clan" && side != "opponent" {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "side must be clan or opponent"))
		return
	}

	pagination := utils.ParsePagination(ctx)
	result, err := c.service.ListMembers(ctx.Request.Context(), snapshotID, side, pagination)
	if err != nil {
		failWar(ctx, err)
		return
	}
	utils.OK(ctx, paginated(result.Items, pagination, result.Total))
}

func failWar(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorInvalidTag:
			utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), warErr.Code, nil)
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorWarNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	if errors.Is(err, wardomain.ErrSnapshotNotFound) {
		utils.Fail(ctx, utils.NewError(utils.ErrNotFound, "war snapshot not found"))
		return
	}
	utils.Fail(ctx, err)
}
