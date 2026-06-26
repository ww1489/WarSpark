package service

import (
	"context"
	"encoding/json"
	"strings"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/capital"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
)

type capitalCache interface {
	GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, bool, error)
	SetCapitalRaidSeasons(ctx context.Context, clanTag string, resp capital.CapitalRaidSeasonListResponse) error
	GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, bool, error)
	SetCapitalLeagues(ctx context.Context, resp capital.CapitalLeagueListResponse) error
	GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, bool, error)
	SetCapitalLeague(ctx context.Context, leagueID string, resp capital.CapitalLeague) error
	GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, bool, error)
	SetBuilderBaseLeagues(ctx context.Context, resp capital.BuilderBaseLeagueListResponse) error
	GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, bool, error)
	SetBuilderBaseLeague(ctx context.Context, leagueID string, resp capital.BuilderBaseLeague) error
}

type cocapiCapitalClient interface {
	GetBuilderBaseLeagues(ctx context.Context, query cocapi.QueryGetBuilderBaseLeagues) (cocapi.BuilderBaseLeagueListResponse, error)
	GetBuilderBaseLeague(ctx context.Context, leagueId string) (cocapi.BuilderBaseLeague, error)
	GetCapitalLeagues(ctx context.Context, query cocapi.QueryGetCapitalLeagues) (cocapi.CapitalLeagueListResponse, error)
	GetCapitalLeague(ctx context.Context, leagueId string) (cocapi.CapitalLeague, error)
	GetCapitalRaidSeasons(ctx context.Context, clanTag string, query cocapi.QueryGetCapitalRaidSeasons) (cocapi.ClanCapitalRaidSeasonsResponse, error)
}

type CapitalService struct {
	cocapi cocapiCapitalClient
	cache  capitalCache
}

func NewCapitalService(cocapi cocapiCapitalClient, cache capitalCache) *CapitalService {
	return &CapitalService{cocapi: cocapi, cache: cache}
}

func (s *CapitalService) GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, error) {
	if clanTag == "" || !strings.HasPrefix(clanTag, "#") {
		return capital.CapitalRaidSeasonListResponse{}, dmerrors.New(dmerrors.ErrCodeInvalidTag, "invalid clan tag")
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetCapitalRaidSeasons(ctx, clanTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetCapitalRaidSeasons(ctx, clanTag, cocapi.QueryGetCapitalRaidSeasons{})
	if err != nil {
		return capital.CapitalRaidSeasonListResponse{}, mapCocapiCapitalError(err)
	}
	resp := capital.CapitalRaidSeasonListResponse{Paging: toDomainCapitalPaging(cocapiResp.Paging)}
	for _, item := range cocapiResp.Items {
		d, err := toDomainCapitalRaidSeason(item)
		if err != nil {
			return capital.CapitalRaidSeasonListResponse{}, err
		}
		resp.Items = append(resp.Items, d)
	}
	if s.cache != nil {
		_ = s.cache.SetCapitalRaidSeasons(ctx, clanTag, resp)
	}
	return resp, nil
}

func (s *CapitalService) GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetCapitalLeagues(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetCapitalLeagues(ctx, cocapi.QueryGetCapitalLeagues{})
	if err != nil {
		return capital.CapitalLeagueListResponse{}, mapCocapiCapitalError(err)
	}
	resp := capital.CapitalLeagueListResponse{Paging: toDomainCapitalPaging(cocapiResp.Paging)}
	for _, item := range cocapiResp.Items {
		resp.Items = append(resp.Items, toDomainCapitalLeague(item))
	}
	if s.cache != nil {
		_ = s.cache.SetCapitalLeagues(ctx, resp)
	}
	return resp, nil
}

func (s *CapitalService) GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, error) {
	if leagueID == "" {
		return capital.CapitalLeague{}, dmerrors.New(dmerrors.ErrCodeInvalidTag, "invalid league id")
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetCapitalLeague(ctx, leagueID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetCapitalLeague(ctx, leagueID)
	if err != nil {
		return capital.CapitalLeague{}, mapCocapiCapitalError(err)
	}
	resp := toDomainCapitalLeague(cocapiResp)
	if s.cache != nil {
		_ = s.cache.SetCapitalLeague(ctx, leagueID, resp)
	}
	return resp, nil
}

func (s *CapitalService) GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetBuilderBaseLeagues(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetBuilderBaseLeagues(ctx, cocapi.QueryGetBuilderBaseLeagues{})
	if err != nil {
		return capital.BuilderBaseLeagueListResponse{}, mapCocapiCapitalError(err)
	}
	resp := capital.BuilderBaseLeagueListResponse{Paging: toDomainCapitalPaging(cocapiResp.Paging)}
	for _, item := range cocapiResp.Items {
		resp.Items = append(resp.Items, toDomainBuilderBaseLeague(item))
	}
	if s.cache != nil {
		_ = s.cache.SetBuilderBaseLeagues(ctx, resp)
	}
	return resp, nil
}

func (s *CapitalService) GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, error) {
	if leagueID == "" {
		return capital.BuilderBaseLeague{}, dmerrors.New(dmerrors.ErrCodeInvalidTag, "invalid league id")
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetBuilderBaseLeague(ctx, leagueID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetBuilderBaseLeague(ctx, leagueID)
	if err != nil {
		return capital.BuilderBaseLeague{}, mapCocapiCapitalError(err)
	}
	resp := toDomainBuilderBaseLeague(cocapiResp)
	if s.cache != nil {
		_ = s.cache.SetBuilderBaseLeague(ctx, leagueID, resp)
	}
	return resp, nil
}

func toDomainCapitalLeague(l cocapi.CapitalLeague) capital.CapitalLeague {
	return capital.CapitalLeague{ID: l.ID, Name: string(l.Name)}
}

func toDomainBuilderBaseLeague(l cocapi.BuilderBaseLeague) capital.BuilderBaseLeague {
	return capital.BuilderBaseLeague{ID: l.ID, Name: string(l.Name)}
}

func toDomainCapitalRaidSeason(s cocapi.ClanCapitalRaidSeason) (capital.CapitalRaidSeason, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return capital.CapitalRaidSeason{}, err
	}
	var d capital.CapitalRaidSeason
	if err := json.Unmarshal(data, &d); err != nil {
		return capital.CapitalRaidSeason{}, err
	}
	return d, nil
}

func toDomainCapitalPaging(p cocapi.Paging) capital.Paging {
	return capital.Paging{Cursors: capital.PagingCursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiCapitalError(err error) error {
	return mapCocapiRankingError(err, "")
}
