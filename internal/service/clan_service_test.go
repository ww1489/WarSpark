package service

import (
	"context"
	"testing"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeClanAPIClient struct {
	detail clandomain.ClanDetail
	err    error
	gotTag string
}

func (f *fakeClanAPIClient) Clan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	f.gotTag = clanTag
	if f.err != nil {
		return clandomain.ClanDetail{}, f.err
	}
	return f.detail, nil
}

func TestClanServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Cached"}}
	cache := &fakeClanCache{clanDetail: cached, clanHit: true}
	svc := NewClanService(&fakeClanAPIClient{}, cache, 5*time.Minute)

	detail, err := svc.FetchClan(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if detail.Clan.Name != "Cached" {
		t.Fatalf("got %q, want Cached", detail.Clan.Name)
	}
}

func TestClanServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakeClanAPIClient{detail: clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Fresh"}}}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache, 5*time.Minute)

	detail, err := svc.FetchClan(context.Background(), "#AAA")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if detail.Clan.Name != "Fresh" {
		t.Fatalf("got %q, want Fresh", detail.Clan.Name)
	}
	if api.gotTag != "#AAA" {
		t.Fatalf("api got tag %q, want #AAA", api.gotTag)
	}
	if cache.setClanTag != "#AAA" {
		t.Fatalf("cache set tag %q, want #AAA", cache.setClanTag)
	}
}

func TestClanServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakeClanAPIClient{err: wardomain.NewError(wardomain.ErrorClanNotFound, "not found")}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache, 5*time.Minute)

	_, err := svc.FetchClan(context.Background(), "#AAA")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorClanNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorClanNotFound)
	}
}

func TestClanServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewClanService(&fakeClanAPIClient{}, &fakeClanCache{}, 5*time.Minute)
	_, err := svc.FetchClan(context.Background(), "!!!")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorInvalidTag {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorInvalidTag)
	}
}

type fakeClanCache struct {
	clanDetail     clandomain.ClanDetail
	clanHit        bool
	setClanTag     string
	playerOverview clandomain.PlayerOverview
	playerHit      bool
	setPlayerTag   string
	battleLog      clandomain.BattleLogSummary
	battleLogHit   bool
	setBattleTag   string
}

func (f *fakeClanCache) GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error) {
	return f.clanDetail, f.clanHit, nil
}
func (f *fakeClanCache) SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error {
	f.setClanTag = clanTag
	return nil
}
func (f *fakeClanCache) GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error) {
	return f.playerOverview, f.playerHit, nil
}
func (f *fakeClanCache) SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error {
	f.setPlayerTag = playerTag
	return nil
}
func (f *fakeClanCache) GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error) {
	return f.battleLog, f.battleLogHit, nil
}
func (f *fakeClanCache) SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error {
	f.setBattleTag = playerTag
	return nil
}
