package service

import (
	"context"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type PlayerAPIClient interface {
	Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error)
	BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error)
}

type PlayerCache interface {
	GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error)
	SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error
	GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error)
	SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error
}

type PlayerService struct {
	client   PlayerAPIClient
	cache    PlayerCache
	cacheTTL time.Duration
}

func NewPlayerService(client PlayerAPIClient, cache PlayerCache, ttl time.Duration) *PlayerService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &PlayerService{client: client, cache: cache, cacheTTL: ttl}
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

	player, err := s.client.Player(ctx, normalizedTag)
	if err != nil {
		return clandomain.PlayerOverview{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetPlayer(ctx, normalizedTag, player, s.cacheTTL)
	}
	return player, nil
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

	log, err := s.client.BattleLog(ctx, normalizedTag)
	if err != nil {
		return clandomain.BattleLogSummary{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetBattleLog(ctx, normalizedTag, log, s.cacheTTL)
	}
	return log, nil
}
