package service

import (
	"context"
	"testing"
	"time"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/domain/league"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeLeagueCache struct {
	leagues       league.LeagueListResponse
	leaguesHit    bool
	l             league.League
	leagueHit     bool
	seasons       league.LeagueSeasonListResponse
	seasonsHit    bool
	rankings      league.LeagueSeasonRankingListResponse
	rankingsHit   bool
	tiers         league.LeagueTierListResponse
	tiersHit      bool
	tier          league.LeagueTier
	tierHit       bool
	history       league.LeagueSeasonResultListResponse
	historyHit    bool
	warLeagues    league.WarLeagueListResponse
	warLeaguesHit bool
	warLeague     league.WarLeague
	warLeagueHit  bool
}

func (f *fakeLeagueCache) GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error) {
	return f.leagues, f.leaguesHit, nil
}
func (f *fakeLeagueCache) SetLeagues(ctx context.Context, resp league.LeagueListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeague(ctx context.Context, id string) (league.League, bool, error) {
	return f.l, f.leagueHit, nil
}
func (f *fakeLeagueCache) SetLeague(ctx context.Context, id string, l league.League, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error) {
	return f.seasons, f.seasonsHit, nil
}
func (f *fakeLeagueCache) SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error) {
	return f.rankings, f.rankingsHit, nil
}
func (f *fakeLeagueCache) SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error) {
	return f.tiers, f.tiersHit, nil
}
func (f *fakeLeagueCache) SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error) {
	return f.tier, f.tierHit, nil
}
func (f *fakeLeagueCache) SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error) {
	return f.history, f.historyHit, nil
}
func (f *fakeLeagueCache) SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error) {
	return f.warLeagues, f.warLeaguesHit, nil
}
func (f *fakeLeagueCache) SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeLeagueCache) GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error) {
	return f.warLeague, f.warLeagueHit, nil
}
func (f *fakeLeagueCache) SetWarLeague(ctx context.Context, id string, l league.WarLeague, ttl time.Duration) error {
	return nil
}

func TestLeagueServiceGetLeaguesCached(t *testing.T) {
	cached := league.LeagueListResponse{Items: []league.League{{ID: 1, Name: "Champion"}}}
	svc := NewLeagueService(&fakeCocapiClient{}, &fakeLeagueCache{leagues: cached, leaguesHit: true}, 30*time.Minute)
	resp, err := svc.GetLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Champion" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeaguesFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagues: cocapi.LeagueListResponse{Items: []cocapi.League{{ID: 1, Name: "Champion"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Champion" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeague(t *testing.T) {
	fakeAPI := &fakeCocapiClient{league: cocapi.League{ID: 1, Name: "Champion"}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	l, err := svc.GetLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if l.Name != "Champion" {
		t.Fatalf("got %q", l.Name)
	}
}

func TestLeagueServiceGetWarLeagues(t *testing.T) {
	fakeAPI := &fakeCocapiClient{warLeagues: cocapi.WarLeagueListResponse{Items: []cocapi.WarLeague{{ID: 1, Name: "Master"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetWarLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Master" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeagueHistory(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueHistory: cocapi.LeagueSeasonResultListResponse{Items: []cocapi.LeagueSeasonResult{{Placement: 1}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagueHistory(context.Background(), "#P1ABC")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Placement != 1 {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceNotFound(t *testing.T) {
	fakeAPI := &fakeCocapiClient{err: cocapi.ErrNotFound}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	_, err := svc.GetLeague(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeLeagueNotFound {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeLeagueNotFound)
	}
}

func TestLeagueServiceGetLeagueSeasons(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueSeasons: cocapi.LeagueSeasonListResponse{Items: []cocapi.LeagueSeason{{ID: "2026-01"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagueSeasons(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID != "2026-01" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeagueSeasonRankings(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueRankings: cocapi.PlayerRankingListResponse{Items: []cocapi.PlayerRanking{{Name: "TopPlayer", Rank: 1}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagueSeasonRankings(context.Background(), "1", "2026-01")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "TopPlayer" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeagueTiers(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueTiers: cocapi.LeagueTierListResponse{Items: []cocapi.LeagueTier{{ID: 1, Name: "Champion I"}}}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	resp, err := svc.GetLeagueTiers(context.Background(), "1", "2026-01")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Champion I" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestLeagueServiceGetLeagueTier(t *testing.T) {
	fakeAPI := &fakeCocapiClient{leagueTier: cocapi.LeagueTier{ID: 1, Name: "Champion I"}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	lt, err := svc.GetLeagueTier(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if lt.Name != "Champion I" {
		t.Fatalf("got %q", lt.Name)
	}
}

func TestLeagueServiceGetWarLeague(t *testing.T) {
	fakeAPI := &fakeCocapiClient{warLeague: cocapi.WarLeague{ID: 1, Name: "Master"}}
	svc := NewLeagueService(fakeAPI, &fakeLeagueCache{}, 30*time.Minute)
	wl, err := svc.GetWarLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if wl.Name != "Master" {
		t.Fatalf("got %q", wl.Name)
	}
}
