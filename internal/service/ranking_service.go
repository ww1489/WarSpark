package service

import (
	"context"
	"errors"
	"time"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/domain/ranking"
)

type rankingCache interface {
	GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error)
	SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error
	GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error)
	SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error
	GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error)
	SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error
	GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error)
	SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error
	GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error)
	SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error)
	SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error
}

type cocapiRankingClient interface {
	GetLocations(ctx context.Context, query cocapi.QueryGetLocations) (cocapi.LocationListResponse, error)
	GetClanRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanRanking) (cocapi.ClanRankingListResponse, error)
	GetPlayerRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerRanking) (cocapi.PlayerRankingListResponse, error)
	GetClanCapitalRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanCapitalRanking) (cocapi.ClanCapitalRankingListResponse, error)
	GetClanBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanBuilderBaseRanking) (cocapi.ClanBuilderBaseRankingListResponse, error)
	GetPlayerBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerBuilderBaseRanking) (cocapi.PlayerBuilderBaseRankingListResponse, error)
}

type RankingService struct {
	cocapi   cocapiRankingClient
	cache    rankingCache
	cacheTTL time.Duration
}

func NewRankingService(cocapi cocapiRankingClient, cache rankingCache, ttl time.Duration) *RankingService {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &RankingService{cocapi: cocapi, cache: cache, cacheTTL: ttl}
}

func (s *RankingService) GetLocations(ctx context.Context) (ranking.LocationListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLocations(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLocations(ctx, cocapi.QueryGetLocations{})
	if err != nil {
		return ranking.LocationListResponse{}, mapCocapiRankingError(err, "")
	}
	resp := ranking.LocationListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, loc := range cocapiResp.Items {
		resp.Items = append(resp.Items, ranking.Location{
			ID: loc.ID, Name: loc.Name, CountryCode: loc.CountryCode,
			IsCountry: loc.IsCountry, LocalizedName: loc.LocalizedName,
		})
	}
	if s.cache != nil {
		_ = s.cache.SetLocations(ctx, resp, s.cacheTTL)
	}
	return resp, nil
}

func (s *RankingService) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanRanking(ctx, locationID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetClanRanking(ctx, locationID, cocapi.QueryGetClanRanking{})
	if err != nil {
		return ranking.ClanRankingListResponse{}, mapCocapiRankingError(err, dmerrors.ErrCodeLocationNotFound)
	}
	resp := ranking.ClanRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanPoints: entry.ClanPoints, Members: entry.Members, Rank: entry.Rank, PreviousRank: entry.PreviousRank, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 {
			item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetClanRanking(ctx, locationID, resp, s.cacheTTL)
	}
	return resp, nil
}

func (s *RankingService) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerRanking(ctx, locationID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetPlayerRanking(ctx, locationID, cocapi.QueryGetPlayerRanking{})
	if err != nil {
		return ranking.PlayerRankingListResponse{}, mapCocapiRankingError(err, dmerrors.ErrCodeLocationNotFound)
	}
	resp := ranking.PlayerRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.PlayerRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, Trophies: entry.Trophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank, AttackWins: entry.AttackWins, DefenseWins: entry.DefenseWins}
		if entry.Clan.Tag != "" {
			item.Clan = &ranking.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetPlayerRanking(ctx, locationID, resp, s.cacheTTL)
	}
	return resp, nil
}

func (s *RankingService) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanCapitalRanking(ctx, locationID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetClanCapitalRanking(ctx, locationID, cocapi.QueryGetClanCapitalRanking{})
	if err != nil {
		return ranking.ClanCapitalRankingListResponse{}, mapCocapiRankingError(err, dmerrors.ErrCodeLocationNotFound)
	}
	resp := ranking.ClanCapitalRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanCapitalRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanCapitalPoints: entry.ClanCapitalPoints, Rank: entry.Rank, PreviousRank: entry.PreviousRank, Members: entry.Members, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 {
			item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetClanCapitalRanking(ctx, locationID, resp, s.cacheTTL)
	}
	return resp, nil
}

func (s *RankingService) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanBuilderBaseRanking(ctx, locationID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetClanBuilderBaseRanking(ctx, locationID, cocapi.QueryGetClanBuilderBaseRanking{})
	if err != nil {
		return ranking.ClanBuilderBaseRankingListResponse{}, mapCocapiRankingError(err, dmerrors.ErrCodeLocationNotFound)
	}
	resp := ranking.ClanBuilderBaseRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.ClanBuilderBaseRankingEntry{Tag: entry.Tag, Name: entry.Name, ClanLevel: entry.ClanLevel, ClanBuilderBasePoints: entry.ClanBuilderBasePoints, Rank: entry.Rank, PreviousRank: entry.PreviousRank, Members: entry.Members, BadgeURLs: entry.BadgeURLs}
		if entry.Location.ID != 0 {
			item.Location = &ranking.Location{ID: entry.Location.ID, Name: entry.Location.Name, CountryCode: entry.Location.CountryCode, IsCountry: entry.Location.IsCountry, LocalizedName: entry.Location.LocalizedName}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetClanBuilderBaseRanking(ctx, locationID, resp, s.cacheTTL)
	}
	return resp, nil
}

func (s *RankingService) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerBuilderBaseRanking(ctx, locationID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetPlayerBuilderBaseRanking(ctx, locationID, cocapi.QueryGetPlayerBuilderBaseRanking{})
	if err != nil {
		return ranking.PlayerBuilderBaseRankingListResponse{}, mapCocapiRankingError(err, dmerrors.ErrCodeLocationNotFound)
	}
	resp := ranking.PlayerBuilderBaseRankingListResponse{Paging: toDomainRankingPaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := ranking.PlayerBuilderBaseRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, BuilderBaseTrophies: entry.BuilderBaseTrophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank}
		if entry.Clan.Tag != "" {
			item.Clan = &ranking.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetPlayerBuilderBaseRanking(ctx, locationID, resp, s.cacheTTL)
	}
	return resp, nil
}

func toDomainRankingPaging(p cocapi.Paging) ranking.Paging {
	return ranking.Paging{Cursors: ranking.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiRankingError(err error, notFoundCode string) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		if notFoundCode != "" {
			return dmerrors.New(notFoundCode, err.Error())
		}
		return err
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return dmerrors.New(dmerrors.ErrCodeAPINotConfigured, err.Error())
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return dmerrors.New(dmerrors.ErrCodeAPIAccessDenied, err.Error())
	case errors.Is(err, cocapi.ErrAPIRequestFailed):
		return dmerrors.New(dmerrors.ErrCodeAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrRateLimited):
		return dmerrors.New(dmerrors.ErrCodeAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return dmerrors.New(dmerrors.ErrCodeAPIResponseInvalid, err.Error())
	default:
		return err
	}
}
