package service

import (
	"context"
	"errors"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type cocapiPlayerClient interface {
	GetPlayer(ctx context.Context, playerTag string) (cocapi.Player, error)
	GetBattleLog(ctx context.Context, playerTag string) (cocapi.BattleLogEntryListResponse, error)
}

type PlayerCache interface {
	GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error)
	SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview) error
	GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error)
	SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary) error
}

type PlayerService struct {
	cocapi cocapiPlayerClient
	cache  PlayerCache
}

func NewPlayerService(cocapi cocapiPlayerClient, cache PlayerCache) *PlayerService {
	return &PlayerService{cocapi: cocapi, cache: cache}
}

func (s *PlayerService) FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil {
		return clandomain.PlayerOverview{}, err
	}

	if s.cache != nil {
		player, ok, err := s.cache.GetPlayer(ctx, normalizedTag)
		if err == nil && ok {
			return player, nil
		}
	}

	player, err := s.cocapi.GetPlayer(ctx, normalizedTag)
	if err != nil {
		return clandomain.PlayerOverview{}, mapCocapiPlayerError(err)
	}
	overview := toDomainPlayerOverview(player)
	if s.cache != nil {
		_ = s.cache.SetPlayer(ctx, normalizedTag, overview)
	}
	return overview, nil
}

func (s *PlayerService) FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, err
	}

	if s.cache != nil {
		log, ok, err := s.cache.GetBattleLog(ctx, normalizedTag)
		if err == nil && ok {
			return log, nil
		}
	}

	log, err := s.cocapi.GetBattleLog(ctx, normalizedTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, mapCocapiPlayerError(err)
	}
	summary := toDomainBattleLog(log)
	if s.cache != nil {
		_ = s.cache.SetBattleLog(ctx, normalizedTag, summary)
	}
	return summary, nil
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
	return clandomain.BattleLogSummary{Items: items}
}

func mapCocapiPlayerError(err error) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		return dmerrors.Wrap(dmerrors.ErrCodePlayerNotFound, err.Error(), err)
	case errors.Is(err, cocapi.ErrInvalidTag):
		return dmerrors.Wrap(dmerrors.ErrCodeInvalidTag, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return dmerrors.Wrap(dmerrors.ErrCodeAPINotConfigured, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return dmerrors.Wrap(dmerrors.ErrCodeAPIAccessDenied, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return dmerrors.Wrap(dmerrors.ErrCodeAPIResponseInvalid, err.Error(), err)
	}
	return dmerrors.Wrap(dmerrors.ErrCodeAPIRequestFailed, err.Error(), err)
}
