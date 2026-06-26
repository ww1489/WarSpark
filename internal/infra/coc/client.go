// Package coc 提供 Clash of Clans API 的领域适配层。
//
// 本包是 pkg/cocapi(官方 swagger 类型)与 internal/domain/war(业务领域类型)之间的 adapter,
// 实现 service.WarAPIClient interface。职责:
//   - 调用 pkg/cocapi.Client 获取官方 API 数据
//   - 将 cocapi 类型转换为 wardomain 类型
//   - 将 cocapi 哨兵错误转换为 dmerrors.Error(带 Code,供 controller 做 HTTP 状态码映射)
//
// 通过 adapter 模式隔离官方 API 类型与业务领域模型,二者可独立演进。
package coc

import (
	"context"
	"errors"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
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

func (c *Client) Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	clan, err := c.api.GetClan(ctx, clanTag)
	if err != nil {
		return clandomain.ClanDetail{}, mapErrorWithNotFound(err, dmerrors.ErrCodeClanNotFound)
	}
	membersResp, err := c.api.GetClanMembers(ctx, clanTag, cocapi.QueryGetClanMembers{})
	if err != nil {
		return clandomain.ClanDetail{}, mapErrorWithNotFound(err, dmerrors.ErrCodeClanNotFound)
	}
	return clandomain.ClanDetail{
		Clan:    toDomainClanOverview(clan),
		Members: toDomainClanMembers(membersResp.Items),
	}, nil
}

func (c *Client) Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	player, err := c.api.GetPlayer(ctx, playerTag)
	if err != nil {
		return clandomain.PlayerOverview{}, mapErrorWithNotFound(err, dmerrors.ErrCodePlayerNotFound)
	}
	return toDomainPlayerOverview(player), nil
}

func (c *Client) BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	log, err := c.api.GetBattleLog(ctx, playerTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, mapErrorWithNotFound(err, dmerrors.ErrCodePlayerNotFound)
	}
	return toDomainBattleLog(log), nil
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

// mapError 把 cocapi 哨兵错误转成 dmerrors.Error(带 Code)。
// controller.failWar 依赖 dmerrors.Error.Code 做 HTTP 状态码映射。
var errorMap = []struct {
	src  error
	code string
}{
	{cocapi.ErrAPINotConfigured, dmerrors.ErrCodeAPINotConfigured},
	{cocapi.ErrAPIAccessDenied, dmerrors.ErrCodeAPIAccessDenied},
	{cocapi.ErrNotFound, dmerrors.ErrCodeWarNotFound},
	{cocapi.ErrInvalidTag, dmerrors.ErrCodeInvalidTag},
	{cocapi.ErrAPIResponseInvalid, dmerrors.ErrCodeAPIResponseInvalid},
	{cocapi.ErrAPIRequestFailed, dmerrors.ErrCodeAPIRequestFailed},
	{cocapi.ErrRateLimited, dmerrors.ErrCodeAPIRequestFailed},
}

func mapError(err error) error {
	for _, m := range errorMap {
		if errors.Is(err, m.src) {
			return dmerrors.Wrap(m.code, err.Error(), err)
		}
	}
	return dmerrors.Wrap(dmerrors.ErrCodeAPIRequestFailed, "unexpected coc api error", err)
}

func mapErrorWithNotFound(err error, notFoundCode string) error {
	mapped := mapError(err)
	var warErr dmerrors.Error
	if errors.As(mapped, &warErr) && warErr.Code == dmerrors.ErrCodeWarNotFound {
		return dmerrors.Wrap(notFoundCode, err.Error(), err)
	}
	return mapped
}

func toDomainClanOverview(c cocapi.Clan) clandomain.ClanOverview {
	overview := clandomain.ClanOverview{
		Tag:            c.Tag,
		Name:           c.Name,
		ClanLevel:      c.ClanLevel,
		Description:    c.Description,
		Members:        c.Members,
		ClanPoints:     c.ClanPoints,
		WarWins:        c.WarWins,
		WarLosses:      c.WarLosses,
		WarTies:        c.WarTies,
		WarWinStreak:   c.WarWinStreak,
		WarFrequency:   c.WarFrequency,
		Type:           c.Type,
		IsWarLogPublic: c.IsWarLogPublic,
		BadgeURLs:      c.BadgeURLs,
		Labels:         toDomainLabels(c.Labels),
	}
	if c.Location.ID != 0 {
		loc := toDomainLocation(c.Location)
		overview.Location = &loc
	}
	if c.WarLeague.ID != 0 {
		league := toDomainLeagueRef(cocapi.League{ID: c.WarLeague.ID, Name: c.WarLeague.Name})
		overview.WarLeague = &league
	}
	return overview
}

func toDomainClanMembers(members []cocapi.ClanMember) []clandomain.ClanMemberSummary {
	result := make([]clandomain.ClanMemberSummary, 0, len(members))
	for _, m := range members {
		summary := clandomain.ClanMemberSummary{
			Tag:               m.Tag,
			Name:              m.Name,
			TownHallLevel:     m.TownHallLevel,
			ExpLevel:          m.ExpLevel,
			Role:              m.Role,
			Trophies:          m.Trophies,
			ClanRank:          m.ClanRank,
			Donations:         m.Donations,
			DonationsReceived: m.DonationsReceived,
		}
		if m.League.ID != 0 {
			league := toDomainLeagueRef(cocapi.League(m.League))
			summary.League = &league
		}
		result = append(result, summary)
	}
	return result
}

func toDomainPlayerOverview(p cocapi.Player) clandomain.PlayerOverview {
	overview := clandomain.PlayerOverview{
		Tag:                 p.Tag,
		Name:                p.Name,
		TownHallLevel:       p.TownHallLevel,
		TownHallWeaponLevel: p.TownHallWeaponLevel,
		ExpLevel:            p.ExpLevel,
		Role:                p.Role,
		WarStars:            p.WarStars,
		AttackWins:          p.AttackWins,
		DefenseWins:         p.DefenseWins,
		Trophies:            p.Trophies,
		BestTrophies:        p.BestTrophies,
		WarPreference:       p.WarPreference,
		Labels:              toDomainLabels(p.Labels),
		Heroes:              toDomainHeroLevels(p.Heroes),
		Achievements:        toDomainAchievements(p.Achievements),
	}
	if p.Clan.Tag != "" {
		overview.Clan = &clandomain.PlayerClanInfo{
			Tag:       p.Clan.Tag,
			Name:      p.Clan.Name,
			ClanLevel: p.Clan.ClanLevel,
			BadgeURLs: p.Clan.BadgeURLs,
		}
	}
	if p.League.ID != 0 {
		league := toDomainLeagueRef(p.League)
		overview.League = &league
	}
	return overview
}

func toDomainBattleLog(log cocapi.BattleLogEntryListResponse) clandomain.BattleLogSummary {
	items := make([]clandomain.BattleLogEntry, 0, len(log.Items))
	for _, e := range log.Items {
		items = append(items, clandomain.BattleLogEntry{
			ArmyShareCode:         e.ArmyShareCode,
			Attack:                e.Attack,
			BattleTime:            e.BattleTime,
			BattleTimestamp:       e.BattleTimestamp,
			BattleType:            e.BattleType,
			DestructionPercentage: e.DestructionPercentage,
			OpponentName:          e.OpponentName,
			OpponentPlayerTag:     e.OpponentPlayerTag,
			OpponentTownHallLevel: e.OpponentTownHallLevel,
			Stars:                 e.Stars,
		})
	}
	return clandomain.BattleLogSummary{Items: items, Paging: clandomain.Paging{}}
}

func toDomainLabels(labels cocapi.LabelList) []clandomain.Label {
	result := make([]clandomain.Label, 0, len(labels))
	for _, l := range labels {
		result = append(result, clandomain.Label{
			ID:   l.ID,
			Name: string(l.Name),
		})
	}
	return result
}

func toDomainLocation(loc cocapi.Location) clandomain.Location {
	return clandomain.Location{
		ID:          loc.ID,
		Name:        string(loc.Name),
		CountryCode: loc.CountryCode,
	}
}

func toDomainLeagueRef(league cocapi.League) clandomain.LeagueRef {
	return clandomain.LeagueRef{
		ID:   league.ID,
		Name: string(league.Name),
	}
}

func toDomainHeroLevels(heroes cocapi.PlayerItemLevelList) []clandomain.HeroLevel {
	result := make([]clandomain.HeroLevel, 0, len(heroes))
	for _, h := range heroes {
		result = append(result, clandomain.HeroLevel{
			Name:     string(h.Name),
			Level:    h.Level,
			MaxLevel: h.MaxLevel,
			Village:  h.Village,
		})
	}
	return result
}

func toDomainAchievements(achievements cocapi.PlayerAchievementProgressList) []clandomain.AchievementProgress {
	result := make([]clandomain.AchievementProgress, 0, len(achievements))
	for _, a := range achievements {
		result = append(result, clandomain.AchievementProgress{
			Name:    string(a.Name),
			Stars:   a.Stars,
			Target:  a.Target,
			Value:   a.Value,
			Village: a.Village,
			Info:    string(a.Info),
		})
	}
	return result
}
