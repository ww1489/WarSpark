package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/utility"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type UtilityReader interface {
	GetCurrentGoldPassSeason(ctx context.Context) (utility.GoldPassSeason, error)
	SearchClans(ctx context.Context, params utility.ClanSearchParams) (cocapi.ClanListResponse, error)
	GetLocation(ctx context.Context, locationID string) (cocapi.Location, error)
	VerifyPlayerToken(ctx context.Context, playerTag, token string) (cocapi.VerifyTokenResponse, error)
	GetPlayerLeagueGroup(ctx context.Context) (utility.PlayerLeagueGroup, error)
}

type UtilityController struct {
	service UtilityReader
}

func NewUtilityController(service UtilityReader) *UtilityController {
	return &UtilityController{service: service}
}

// GetCurrentGoldPass 获取当前 Gold Pass 赛季信息。
//
// @Summary 获取当前 Gold Pass
// @Tags utility
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/gold-pass/current [get]
func (ctl *UtilityController) GetCurrentGoldPass(c *gin.Context) {
	resp, err := ctl.service.GetCurrentGoldPassSeason(c.Request.Context())
	if err != nil {
		failUtility(c, err)
		return
	}
	utils.OK(c, resp)
}

// SearchClans 搜索部落。
//
// @Summary 搜索部落
// @Tags utility
// @Produce json
// @Param name query string false "部落名称"
// @Param warFrequency query string false "战争频率"
// @Param locationId query int false "位置ID"
// @Param minMembers query int false "最少成员数"
// @Param maxMembers query int false "最多成员数"
// @Param minClanPoints query int false "最少部落积分"
// @Param minClanLevel query int false "最少部落等级"
// @Param limit query int false "分页数量"
// @Param after query string false "游标后"
// @Param before query string false "游标前"
// @Param labelIds query string false "标签ID列表"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/clans [get]
func (ctl *UtilityController) SearchClans(c *gin.Context) {
	var params utility.ClanSearchParams
	if !utils.BindQuery(c, &params) {
		return
	}
	resp, err := ctl.service.SearchClans(c.Request.Context(), params)
	if err != nil {
		failUtility(c, err)
		return
	}
	utils.OK(c, resp)
}

// GetLocation 获取位置详情。
//
// @Summary 获取位置详情
// @Tags utility
// @Produce json
// @Param id path string true "位置ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/locations/{id} [get]
func (ctl *UtilityController) GetLocation(c *gin.Context) {
	locationID := c.Param("id")
	resp, err := ctl.service.GetLocation(c.Request.Context(), locationID)
	if err != nil {
		failUtility(c, err)
		return
	}
	utils.OK(c, resp)
}

// VerifyPlayerToken 验证玩家 API 令牌。
//
// @Summary 验证玩家令牌
// @Tags utility
// @Produce json
// @Param tag path string true "玩家标签"
// @Param body body utility.VerifyTokenRequest true "验证请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 502 {object} utils.Response
// @Router /api/v1/players/{tag}/verify-token [post]
func (ctl *UtilityController) VerifyPlayerToken(c *gin.Context) {
	var body utility.VerifyTokenRequest
	if !utils.BindJSON(c, &body) {
		return
	}
	playerTag := c.Param("tag")
	resp, err := ctl.service.VerifyPlayerToken(c.Request.Context(), playerTag, body.Token)
	if err != nil {
		failUtility(c, err)
		return
	}
	utils.OK(c, resp)
}

// GetPlayerLeagueGroup 获取玩家联赛分组（暂未实现）。
//
// @Summary 获取玩家联赛分组
// @Tags utility
// @Produce json
// @Param tag path string true "玩家标签"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/players/{tag}/league-group [get]
func (ctl *UtilityController) GetPlayerLeagueGroup(c *gin.Context) {
	_, err := ctl.service.GetPlayerLeagueGroup(c.Request.Context())
	if err != nil {
		failUtility(c, err)
		return
	}
	utils.OK(c, nil)
}

func failUtility(ctx *gin.Context, err error) {
	var warErr wardomain.Error
	if errors.As(err, &warErr) {
		switch warErr.Code {
		case wardomain.ErrorAPINotConfigured:
			utils.JSON(ctx, http.StatusServiceUnavailable, int(utils.ErrInternal), warErr.Code, nil)
		case wardomain.ErrorAPIAccessDenied:
			utils.JSON(ctx, http.StatusForbidden, int(utils.ErrForbidden), warErr.Code, nil)
		case "not_implemented":
			utils.JSON(ctx, http.StatusBadRequest, int(utils.ErrInvalidField), warErr.Code, nil)
		default:
			utils.JSON(ctx, http.StatusBadGateway, int(utils.ErrInternal), warErr.Code, nil)
		}
		return
	}
	utils.Fail(ctx, err)
}
