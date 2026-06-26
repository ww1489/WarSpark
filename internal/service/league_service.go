package service

import (
	"context"
	"errors"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/domain/league"
)

type cocapiLeagueClient interface {
	GetLeagues(ctx context.Context, query cocapi.QueryGetLeagues) (cocapi.LeagueListResponse, error)
	GetLeague(ctx context.Context, leagueID string) (cocapi.League, error)
	GetLeagueSeasons(ctx context.Context, leagueID string, query cocapi.QueryGetLeagueSeasons) (cocapi.LeagueSeasonListResponse, error)
	GetLeagueSeasonRankings(ctx context.Context, leagueID, seasonID string, query cocapi.QueryGetLeagueSeasonRankings) (cocapi.PlayerRankingListResponse, error)
	GetLeagueTiers(ctx context.Context, query cocapi.QueryGetLeagueTiers) (cocapi.LeagueTierListResponse, error)
	GetLeagueTier(ctx context.Context, leagueTierID string) (cocapi.LeagueTier, error)
	GetLeagueHistory(ctx context.Context, playerTag string) (cocapi.LeagueSeasonResultListResponse, error)
	GetWarLeagues(ctx context.Context, query cocapi.QueryGetWarLeagues) (cocapi.WarLeagueListResponse, error)
	GetWarLeague(ctx context.Context, leagueID string) (cocapi.WarLeague, error)
}

type leagueCache interface {
	GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error)
	SetLeagues(ctx context.Context, resp league.LeagueListResponse) error
	GetLeague(ctx context.Context, id string) (league.League, bool, error)
	SetLeague(ctx context.Context, id string, l league.League) error
	GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error)
	SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse) error
	GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error)
	SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse) error
	GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error)
	SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse) error
	GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error)
	SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier) error
	GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error)
	SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse) error
	GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error)
	SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse) error
	GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error)
	SetWarLeague(ctx context.Context, id string, l league.WarLeague) error
}

type LeagueService struct {
	cocapi cocapiLeagueClient
	cache  leagueCache
}

func NewLeagueService(cocapi cocapiLeagueClient, cache leagueCache) *LeagueService {
	return &LeagueService{cocapi: cocapi, cache: cache}
}

func (s *LeagueService) GetLeagues(ctx context.Context) (league.LeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagues(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLeagues(ctx, cocapi.QueryGetLeagues{})
	if err != nil {
		return league.LeagueListResponse{}, mapCocapiLeagueError(err, "")
	}
	resp := league.LeagueListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, l := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.League{ID: l.ID, Name: string(l.Name), IconURLs: l.IconURLs})
	}
	if s.cache != nil {
		_ = s.cache.SetLeagues(ctx, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetLeague(ctx context.Context, leagueID string) (league.League, error) {
	if s.cache != nil {
		l, ok, err := s.cache.GetLeague(ctx, leagueID)
		if err == nil && ok {
			return l, nil
		}
	}
	l, err := s.cocapi.GetLeague(ctx, leagueID)
	if err != nil {
		return league.League{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	dom := league.League{ID: l.ID, Name: string(l.Name), IconURLs: l.IconURLs}
	if s.cache != nil {
		_ = s.cache.SetLeague(ctx, leagueID, dom)
	}
	return dom, nil
}

func (s *LeagueService) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueSeasons(ctx, leagueID)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLeagueSeasons(ctx, leagueID, cocapi.QueryGetLeagueSeasons{})
	if err != nil {
		return league.LeagueSeasonListResponse{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	resp := league.LeagueSeasonListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, s := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.LeagueSeason{ID: s.ID})
	}
	if s.cache != nil {
		_ = s.cache.SetLeagueSeasons(ctx, leagueID, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueSeasonRankings(ctx, leagueID, season)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLeagueSeasonRankings(ctx, leagueID, season, cocapi.QueryGetLeagueSeasonRankings{})
	if err != nil {
		return league.LeagueSeasonRankingListResponse{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	resp := league.LeagueSeasonRankingListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, entry := range cocapiResp.Items {
		item := league.LeagueSeasonRankingEntry{Tag: entry.Tag, Name: entry.Name, ExpLevel: entry.ExpLevel, Trophies: entry.Trophies, Rank: entry.Rank, PreviousRank: entry.PreviousRank, AttackWins: entry.AttackWins, DefenseWins: entry.DefenseWins}
		if entry.Clan.Tag != "" {
			item.Clan = &league.ClanRef{Tag: entry.Clan.Tag, Name: entry.Clan.Name, BadgeURLs: entry.Clan.BadgeURLs}
		}
		resp.Items = append(resp.Items, item)
	}
	if s.cache != nil {
		_ = s.cache.SetLeagueSeasonRankings(ctx, leagueID, season, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueTiers(ctx, leagueID, season)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLeagueTiers(ctx, cocapi.QueryGetLeagueTiers{})
	if err != nil {
		return league.LeagueTierListResponse{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	resp := league.LeagueTierListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, t := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.LeagueTier{ID: t.ID, Name: string(t.Name), IconURLs: t.IconURLs})
	}
	if s.cache != nil {
		_ = s.cache.SetLeagueTiers(ctx, leagueID, season, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error) {
	if s.cache != nil {
		t, ok, err := s.cache.GetLeagueTier(ctx, tierID)
		if err == nil && ok {
			return t, nil
		}
	}
	t, err := s.cocapi.GetLeagueTier(ctx, tierID)
	if err != nil {
		return league.LeagueTier{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	dom := league.LeagueTier{ID: t.ID, Name: string(t.Name), IconURLs: t.IconURLs}
	if s.cache != nil {
		_ = s.cache.SetLeagueTier(ctx, tierID, dom)
	}
	return dom, nil
}

func (s *LeagueService) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error) {
	normalizedTag, err := NormalizeClanTag(playerTag)
	if err != nil {
		return league.LeagueSeasonResultListResponse{}, err
	}
	if s.cache != nil {
		resp, ok, err := s.cache.GetLeagueHistory(ctx, normalizedTag)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetLeagueHistory(ctx, normalizedTag)
	if err != nil {
		return league.LeagueSeasonResultListResponse{}, mapCocapiLeagueError(err, "")
	}
	resp := league.LeagueSeasonResultListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, r := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.LeagueSeasonResult{
			Trophies:  r.LeagueTrophies,
			Placement: r.Placement, AttackWins: r.AttackWins, AttackLosses: r.AttackLosses, AttackStars: r.AttackStars,
			DefenseWins: r.DefenseWins, DefenseLosses: r.DefenseLosses, DefenseStars: r.DefenseStars, MaxBattles: r.MaxBattles,
		})
	}
	if s.cache != nil {
		_ = s.cache.SetLeagueHistory(ctx, normalizedTag, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetWarLeagues(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetWarLeagues(ctx, cocapi.QueryGetWarLeagues{})
	if err != nil {
		return league.WarLeagueListResponse{}, mapCocapiLeagueError(err, "")
	}
	resp := league.WarLeagueListResponse{Paging: toDomainLeaguePaging(cocapiResp.Paging)}
	for _, wl := range cocapiResp.Items {
		resp.Items = append(resp.Items, league.WarLeague{ID: wl.ID, Name: string(wl.Name)})
	}
	if s.cache != nil {
		_ = s.cache.SetWarLeagues(ctx, resp)
	}
	return resp, nil
}

func (s *LeagueService) GetWarLeague(ctx context.Context, leagueID string) (league.WarLeague, error) {
	if s.cache != nil {
		l, ok, err := s.cache.GetWarLeague(ctx, leagueID)
		if err == nil && ok {
			return l, nil
		}
	}
	wl, err := s.cocapi.GetWarLeague(ctx, leagueID)
	if err != nil {
		return league.WarLeague{}, mapCocapiLeagueError(err, dmerrors.ErrCodeLeagueNotFound)
	}
	dom := league.WarLeague{ID: wl.ID, Name: string(wl.Name)}
	if s.cache != nil {
		_ = s.cache.SetWarLeague(ctx, leagueID, dom)
	}
	return dom, nil
}

func toDomainLeaguePaging(p cocapi.Paging) league.Paging {
	return league.Paging{Cursors: league.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiLeagueError(err error, notFoundCode string) error {
	if notFoundCode != "" && errors.Is(err, cocapi.ErrNotFound) {
		return dmerrors.New(notFoundCode, err.Error())
	}
	return mapCocapiRankingError(err, notFoundCode)
}
