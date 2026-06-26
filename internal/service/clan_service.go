package service

import (
	"context"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type ClanAPIClient interface {
	Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error)
}

type ClanCache interface {
	GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error)
	SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error
}

type ClanService struct {
	client   ClanAPIClient
	cache    ClanCache
	cacheTTL time.Duration
}

func NewClanService(client ClanAPIClient, cache ClanCache, ttl time.Duration) *ClanService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &ClanService{client: client, cache: cache, cacheTTL: ttl}
}

func (s *ClanService) FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return clandomain.ClanDetail{}, err
	}

	if s.cache != nil {
		detail, ok, err := s.cache.GetClan(ctx, normalizedTag)
		if err == nil && ok {
			return detail, nil
		}
	}

	detail, err := s.client.Clan(ctx, normalizedTag)
	if err != nil {
		return clandomain.ClanDetail{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetClan(ctx, normalizedTag, detail, s.cacheTTL)
	}
	return detail, nil
}
