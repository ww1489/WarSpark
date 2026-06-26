package coc

import (
	"context"
	"errors"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type Client struct {
	api *cocapi.Client
}

func New(cfg cocapi.Config) *Client {
	return &Client{api: cocapi.New(cfg)}
}

func (c *Client) CurrentWar(ctx context.Context, clanTag string) (wardomain.CurrentWar, error) {
	cw, err := c.api.GetCurrentWar(ctx, clanTag)
	if err != nil {
		return wardomain.CurrentWar{}, mapError(err)
	}
	return toDomainCurrentWar(cw), nil
}

func (c *Client) CWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, error) {
	group, err := c.api.GetClanWarLeagueGroup(ctx, clanTag)
	if err != nil {
		return wardomain.CWLGroup{}, mapError(err)
	}
	return toDomainCWLGroup(group), nil
}

func toDomainCurrentWar(cw cocapi.ClanWar) wardomain.CurrentWar {
	return wardomain.CurrentWar{
		State:    cw.State,
		TeamSize: cw.TeamSize,
		Clan:     toDomainWarClan(cw.Clan),
		Opponent: toDomainWarClan(cw.Opponent),
	}
}

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
	var appErr dmerrors.Error
	if errors.As(mapped, &appErr) && appErr.Code == dmerrors.ErrCodeWarNotFound {
		return dmerrors.Wrap(notFoundCode, err.Error(), err)
	}
	return mapped
}
