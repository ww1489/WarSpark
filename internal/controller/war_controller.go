package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type WarReader interface {
	FetchCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, error)
	ListMembers(ctx context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error)
	FetchCWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error)
	GetWarLog(ctx context.Context, clanTag string, limit int, after, before string) (cocapi.ClanWarLogResponse, error)
	GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, error)
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

// GetCWL returns the current Clan War League group for a clan.
//
// @Summary Get CWL group
// @Tags war
// @Produce json
// @Param clan_tag query string true "Clan tag"
// @Success 200 {object} utils.Response
// @Failure 422 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/war/cwl [get]
func (c *WarController) GetCWL(ctx *gin.Context) {
	clanTag := ctx.Query("clan_tag")
	if clanTag == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "clan_tag is required"))
		return
	}

	group, err := c.service.FetchCWLGroup(ctx.Request.Context(), clanTag)
	if err != nil {
		failWar(ctx, err)
		return
	}
	utils.OK(ctx, group)
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

// GetWarLog 获取部落战争日志。
//
// @Summary 获取部落战争日志
// @Tags war
// @Produce json
// @Param tag path string true "部落标签"
// @Param limit query int false "分页数量"
// @Param after query string false "游标后"
// @Param before query string false "游标前"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/clans/{tag}/war-log [get]
func (ctl *WarController) GetWarLog(c *gin.Context) {
	tag := c.Param("tag")
	var query struct {
		Limit  int    `form:"limit"`
		After  string `form:"after"`
		Before string `form:"before"`
	}
	if !utils.BindQuery(c, &query) {
		return
	}
	resp, err := ctl.service.GetWarLog(c.Request.Context(), tag, query.Limit, query.After, query.Before)
	if err != nil {
		failWar(c, err)
		return
	}
	utils.OK(c, resp)
}

// GetCWLWar 获取 CWL 联赛单场战争详情。
//
// @Summary 获取 CWL 联赛战争
// @Tags war
// @Produce json
// @Param war_tag path string true "战争标签"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/war/cwl/wars/{war_tag} [get]
func (ctl *WarController) GetCWLWar(c *gin.Context) {
	warTag := c.Param("war_tag")
	resp, err := ctl.service.GetCWLWar(c.Request.Context(), warTag)
	if err != nil {
		failWar(c, err)
		return
	}
	utils.OK(c, resp)
}

func failWar(ctx *gin.Context, err error) {
	var warErr dmerrors.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case dmerrors.ErrCodeInvalidTag:
			utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), warErr.Code, nil)
		case dmerrors.ErrCodeAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case dmerrors.ErrCodeWarNotFound:
			utils.JSON(ctx, http.StatusNotFound, int(utils.ErrNotFound), warErr.Code, nil)
		case dmerrors.ErrCodeAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	if errors.Is(err, dmerrors.ErrSnapshotNotFound) {
		utils.Fail(ctx, utils.NewError(utils.ErrNotFound, "war snapshot not found"))
		return
	}
	utils.Fail(ctx, err)
}
