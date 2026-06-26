package service

import (
	"context"
	"testing"
	"time"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeRankingCache struct {
	locations        ranking.LocationListResponse
	locationsHit     bool
	clanRanking      ranking.ClanRankingListResponse
	clanRankingHit   bool
	playerRanking    ranking.PlayerRankingListResponse
	playerRankingHit bool
}

func (f *fakeRankingCache) GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error) {
	return f.locations, f.locationsHit, nil
}
func (f *fakeRankingCache) SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error) {
	return f.clanRanking, f.clanRankingHit, nil
}
func (f *fakeRankingCache) SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error) {
	return f.playerRanking, f.playerRankingHit, nil
}
func (f *fakeRankingCache) SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error) {
	return ranking.ClanCapitalRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error) {
	return ranking.ClanBuilderBaseRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error { return nil }
func (f *fakeRankingCache) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error) {
	return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil
}
func (f *fakeRankingCache) SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error { return nil }

func TestRankingServiceGetLocationsCached(t *testing.T) {
	cached := ranking.LocationListResponse{Items: []ranking.Location{{ID: 1, Name: "Global"}}}
	svc := NewRankingService(&fakeCocapiClient{}, &fakeRankingCache{locations: cached, locationsHit: true}, 10*time.Minute)
	resp, err := svc.GetLocations(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Global" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetLocationsFromAPI(t *testing.T) {
	fakeAPI := &fakeCocapiClient{
		locations: cocapi.LocationListResponse{Items: []cocapi.Location{{ID: 1, Name: "Global"}}},
	}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetLocations(context.Background())
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "Global" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetClanRanking(t *testing.T) {
	fakeAPI := &fakeCocapiClient{clanRanking: cocapi.ClanRankingListResponse{Items: []cocapi.ClanRanking{{Name: "TopClan", Rank: 1}}}}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetClanRanking(context.Background(), "global")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "TopClan" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceGetPlayerRanking(t *testing.T) {
	fakeAPI := &fakeCocapiClient{playerRanking: cocapi.PlayerRankingListResponse{Items: []cocapi.PlayerRanking{{Name: "TopPlayer", Rank: 1}}}}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	resp, err := svc.GetPlayerRanking(context.Background(), "global")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(resp.Items) != 1 || resp.Items[0].Name != "TopPlayer" {
		t.Fatalf("got %+v", resp.Items)
	}
}

func TestRankingServiceNotFound(t *testing.T) {
	fakeAPI := &fakeCocapiClient{err: cocapi.ErrNotFound}
	svc := NewRankingService(fakeAPI, &fakeRankingCache{}, 10*time.Minute)
	_, err := svc.GetClanRanking(context.Background(), "999")
	if err == nil { t.Fatal("expected error") }
	if wardomain.ErrorCode(err) != wardomain.ErrorLocationNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorLocationNotFound)
	}
}
