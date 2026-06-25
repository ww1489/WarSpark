// Package coc 提供 Clash of Clans API 的领域适配层。
//
// 本包是 pkg/cocapi(官方 swagger 类型)与 internal/domain/war(业务领域类型)之间的 adapter,
// 实现 service.WarAPIClient interface。职责:
//   - 调用 pkg/cocapi.Client 获取官方 API 数据
//   - 将 cocapi 类型转换为 wardomain 类型
//   - 将 cocapi 哨兵错误转换为 wardomain.Error(带 Code,供 controller 做 HTTP 状态码映射)
//
// 通过 adapter 模式隔离官方 API 类型与业务领域模型,二者可独立演进。
package coc

import (
	"context"
	"errors"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

// Client 是 service.WarAPIClient 的实现,内部包装 cocapi.Client。
type Client struct {
	api *cocapi.Client
}

// New 从 cocapi.Config 构造 adapter。
func New(cfg cocapi.Config) *Client {
	return &Client{api: cocapi.New(cfg)}
}

// CurrentWar 获取部族当前战争,返回 wardomain.CurrentWar。
func (c *Client) CurrentWar(ctx context.Context, clanTag string) (wardomain.CurrentWar, error) {
	cw, err := c.api.GetCurrentWar(ctx, clanTag)
	if err != nil {
		return wardomain.CurrentWar{}, mapError(err)
	}
	return toDomainCurrentWar(cw), nil
}

// CWLGroup 获取部族当前 CWL 分组,返回 wardomain.CWLGroup。
// ClanTag/ClanName 留空,由 service 层填充。
func (c *Client) CWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error) {
	group, err := c.api.GetClanWarLeagueGroup(ctx, clanTag)
	if err != nil {
		return wardomain.CWLGroup{}, mapError(err)
	}
	return toDomainCWLGroup(group), nil
}

// toDomainCurrentWar 把 cocapi.ClanWar 转成 wardomain.CurrentWar。
func toDomainCurrentWar(cw cocapi.ClanWar) wardomain.CurrentWar {
	return wardomain.CurrentWar{
		State:    cw.State,
		TeamSize: cw.TeamSize,
		Clan:     toDomainWarClan(cw.Clan),
		Opponent: toDomainWarClan(cw.Opponent),
	}
}

// toDomainWarClan 把 cocapi.WarClan 转成 wardomain.WarClan。
func toDomainWarClan(wc cocapi.WarClan) wardomain.WarClan {
	members := make([]wardomain.WarMember, 0, len(wc.Members))
	for _, m := range wc.Members {
		members = append(members, toDomainWarMember(m))
	}
	return wardomain.WarClan{
		Tag:                   wc.Tag,
		Name:                  wc.Name,
		Stars:                 wc.Stars,
		DestructionPercentage: float64(wc.DestructionPercentage),
		Members:               members,
	}
}

// toDomainWarMember 把 cocapi.ClanWarMember 转成 wardomain.WarMember。
func toDomainWarMember(m cocapi.ClanWarMember) wardomain.WarMember {
	attacks := make([]wardomain.WarAttack, 0, len(m.Attacks))
	for _, a := range m.Attacks {
		attacks = append(attacks, toDomainWarAttack(a))
	}
	return wardomain.WarMember{
		Tag:           m.Tag,
		Name:          m.Name,
		TownHallLevel: m.TownhallLevel,
		MapPosition:   m.MapPosition,
		Attacks:       attacks,
	}
}

// toDomainWarAttack 把 cocapi.ClanWarAttack 转成 wardomain.WarAttack。
func toDomainWarAttack(a cocapi.ClanWarAttack) wardomain.WarAttack {
	return wardomain.WarAttack{
		AttackerTag:           a.AttackerTag,
		DefenderTag:           a.DefenderTag,
		Stars:                 a.Stars,
		DestructionPercentage: float64(a.DestructionPercentage),
		Order:                 a.Order,
		Duration:              a.Duration,
	}
}

// toDomainCWLGroup 把 cocapi.ClanWarLeagueGroup 转成 wardomain.CWLGroup。
// ClanTag/ClanName 留空(service 层负责填充)。
func toDomainCWLGroup(g cocapi.ClanWarLeagueGroup) wardomain.CWLGroup {
	clans := make([]wardomain.CWLClan, 0, len(g.Clans))
	for _, c := range g.Clans {
		clans = append(clans, wardomain.CWLClan{
			Tag:       c.Tag,
			Name:      c.Name,
			ClanLevel: c.ClanLevel,
			Members:   len(c.Members),
		})
	}
	rounds := make([]wardomain.CWLRound, 0, len(g.Rounds))
	for _, r := range g.Rounds {
		rounds = append(rounds, wardomain.CWLRound{
			WarTags: r.WarTags,
		})
	}
	return wardomain.CWLGroup{
		State:  g.State,
		Season: g.Season,
		Clans:  clans,
		Rounds: rounds,
	}
}

// mapError 把 cocapi 哨兵错误转成 wardomain.Error(带 Code)。
// controller.failWar 依赖 wardomain.Error.Code 做 HTTP 状态码映射。
var errorMap = []struct {
	src  error
	code string
}{
	{cocapi.ErrAPINotConfigured, wardomain.ErrorAPINotConfigured},
	{cocapi.ErrAPIAccessDenied, wardomain.ErrorAPIAccessDenied},
	{cocapi.ErrNotFound, wardomain.ErrorWarNotFound},
	{cocapi.ErrInvalidTag, wardomain.ErrorInvalidTag},
	{cocapi.ErrAPIResponseInvalid, wardomain.ErrorAPIResponseInvalid},
	{cocapi.ErrAPIRequestFailed, wardomain.ErrorAPIRequestFailed},
	{cocapi.ErrRateLimited, wardomain.ErrorAPIRequestFailed},
}

func mapError(err error) error {
	for _, m := range errorMap {
		if errors.Is(err, m.src) {
			return wardomain.WrapError(m.code, err.Error(), err)
		}
	}
	return wardomain.WrapError(wardomain.ErrorAPIRequestFailed, "unexpected coc api error", err)
}
