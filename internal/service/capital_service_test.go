package service

import (
	"context"
	"testing"
	"time"

	"github.com/ww1489/WarSpark/internal/domain/capital"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeCapitalCache struct {
	raidSeasons   capital.CapitalRaidSeasonListResponse
	raidSeasonsOK bool
	capLeagues    capital.CapitalLeagueListResponse
	capLeaguesOK  bool
	capLeague     capital.CapitalLeague
	capLeagueOK   bool
	builderLeagues capital.BuilderBaseLeagueListResponse
	builderLeaguesOK bool
	builderLeague capital.BuilderBaseLeague
	builderLeagueOK bool
}

func (f *fakeCapitalCache) GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, bool, error) {
	return f.raidSeasons, f.raidSeasonsOK, nil
}
func (f *fakeCapitalCache) SetCapitalRaidSeasons(ctx context.Context, clanTag string, resp capital.CapitalRaidSeasonListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeCapitalCache) GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, bool, error) {
	return f.capLeagues, f.capLeaguesOK, nil
}
func (f *fakeCapitalCache) SetCapitalLeagues(ctx context.Context, resp capital.CapitalLeagueListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeCapitalCache) GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, bool, error) {
	return f.capLeague, f.capLeagueOK, nil
}
func (f *fakeCapitalCache) SetCapitalLeague(ctx context.Context, leagueID string, resp capital.CapitalLeague, ttl time.Duration) error {
	return nil
}
func (f *fakeCapitalCache) GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, bool, error) {
	return f.builderLeagues, f.builderLeaguesOK, nil
}
func (f *fakeCapitalCache) SetBuilderBaseLeagues(ctx context.Context, resp capital.BuilderBaseLeagueListResponse, ttl time.Duration) error {
	return nil
}
func (f *fakeCapitalCache) GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, bool, error) {
	return f.builderLeague, f.builderLeagueOK, nil
}
func (f *fakeCapitalCache) SetBuilderBaseLeague(ctx context.Context, leagueID string, resp capital.BuilderBaseLeague, ttl time.Duration) error {
	return nil
}

func TestCapitalServiceRaidSeasonsCached(t *testing.T) {
	cached := capital.CapitalRaidSeasonListResponse{
		Items: []capital.CapitalRaidSeason{{State: "inactive"}},
	}
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{raidSeasons: cached, raidSeasonsOK: true}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalRaidSeasons(context.Background(), "#2PP")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].State != "inactive" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceRaidSeasonsFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		capitalRaidSeasons: cocapi.ClanCapitalRaidSeasonsResponse{
			Items: []cocapi.ClanCapitalRaidSeason{{State: "active"}},
		},
	}
	svc := NewCapitalService(fakeAPI, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalRaidSeasons(context.Background(), "#2PP")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].State != "active" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceRaidSeasonsInvalidTag(t *testing.T) {
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	_, err := svc.GetCapitalRaidSeasons(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestCapitalServiceGetCapitalLeaguesCached(t *testing.T) {
	cached := capital.CapitalLeagueListResponse{
		Items: []capital.CapitalLeague{{ID: 1, Name: "Bronze"}},
	}
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{capLeagues: cached, capLeaguesOK: true}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Bronze" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceGetCapitalLeaguesFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		capitalLeagues: cocapi.CapitalLeagueListResponse{
			Items: []cocapi.CapitalLeague{{ID: 1, Name: "Bronze"}},
		},
	}
	svc := NewCapitalService(fakeAPI, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Bronze" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceGetCapitalLeagueCached(t *testing.T) {
	cached := capital.CapitalLeague{ID: 1, Name: "Bronze III"}
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{capLeague: cached, capLeagueOK: true}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Bronze III" {
		t.Fatalf("got %+v", resp)
	}
}

func TestCapitalServiceGetCapitalLeagueFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		capitalLeague: cocapi.CapitalLeague{ID: 1, Name: "Bronze III"},
	}
	svc := NewCapitalService(fakeAPI, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetCapitalLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Bronze III" {
		t.Fatalf("got %+v", resp)
	}
}

func TestCapitalServiceGetCapitalLeagueInvalidID(t *testing.T) {
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	_, err := svc.GetCapitalLeague(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty league id")
	}
}

func TestCapitalServiceGetBuilderBaseLeaguesCached(t *testing.T) {
	cached := capital.BuilderBaseLeagueListResponse{
		Items: []capital.BuilderBaseLeague{{ID: 1, Name: "Wood"}},
	}
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{builderLeagues: cached, builderLeaguesOK: true}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetBuilderBaseLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Wood" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceGetBuilderBaseLeaguesFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		builderBaseLeagues: cocapi.BuilderBaseLeagueListResponse{
			Items: []cocapi.BuilderBaseLeague{{ID: 1, Name: "Wood"}},
		},
	}
	svc := NewCapitalService(fakeAPI, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetBuilderBaseLeagues(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Wood" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestCapitalServiceGetBuilderBaseLeagueCached(t *testing.T) {
	cached := capital.BuilderBaseLeague{ID: 1, Name: "Wood III"}
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{builderLeague: cached, builderLeagueOK: true}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetBuilderBaseLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Wood III" {
		t.Fatalf("got %+v", resp)
	}
}

func TestCapitalServiceGetBuilderBaseLeagueFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		builderBaseLeague: cocapi.BuilderBaseLeague{ID: 1, Name: "Wood III"},
	}
	svc := NewCapitalService(fakeAPI, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	resp, err := svc.GetBuilderBaseLeague(context.Background(), "1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Wood III" {
		t.Fatalf("got %+v", resp)
	}
}

func TestCapitalServiceGetBuilderBaseLeagueInvalidID(t *testing.T) {
	svc := NewCapitalService(&fakeCocapiClient{}, &fakeCapitalCache{}, 5*time.Minute, 30*time.Minute)
	_, err := svc.GetBuilderBaseLeague(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty league id")
	}
}
