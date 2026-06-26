package service

import (
	"context"
	"testing"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeCocapiClanClient struct {
	clan    cocapi.Clan
	members cocapi.ClanMemberListResponse
	err     error
	gotTag  string
}

func (f *fakeCocapiClanClient) GetClan(ctx context.Context, clanTag string) (cocapi.Clan, error) {
	f.gotTag = clanTag
	return f.clan, f.err
}

func (f *fakeCocapiClanClient) GetClanMembers(ctx context.Context, clanTag string, query cocapi.QueryGetClanMembers) (cocapi.ClanMemberListResponse, error) {
	f.gotTag = clanTag
	return f.members, f.err
}

type fakeClanCache struct {
	clanDetail     clandomain.ClanDetail
	clanHit        bool
	playerOverview clandomain.PlayerOverview
	playerHit      bool
	battleLog      clandomain.BattleLogSummary
	battleLogHit   bool
	setDetail      clandomain.ClanDetail
	setTag         string
}

func (f *fakeClanCache) GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error) {
	return f.clanDetail, f.clanHit, nil
}

func (f *fakeClanCache) SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail) error {
	f.setTag = clanTag
	f.setDetail = detail
	return nil
}

func (f *fakeClanCache) GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error) {
	return f.playerOverview, f.playerHit, nil
}

func (f *fakeClanCache) SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview) error {
	return nil
}

func (f *fakeClanCache) GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error) {
	return f.battleLog, f.battleLogHit, nil
}

func (f *fakeClanCache) SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary) error {
	return nil
}

func TestClanServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Cached"}}
	cache := &fakeClanCache{clanDetail: cached, clanHit: true}
	svc := NewClanService(&fakeCocapiClanClient{}, cache)

	result, err := svc.FetchClan(context.Background(), "#AAA111")
	if err != nil {
		t.Fatalf("FetchClan error: %v", err)
	}
	if result.Clan.Name != "Cached" {
		t.Fatalf("got %q, want Cached", result.Clan.Name)
	}
}

func TestClanServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakeCocapiClanClient{
		clan: cocapi.Clan{Name: "TestClan", Tag: "#AAA111", ClanLevel: 10},
	}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache)

	result, err := svc.FetchClan(context.Background(), "#AAA111")
	if err != nil {
		t.Fatalf("FetchClan error: %v", err)
	}
	if result.Clan.Name != "TestClan" {
		t.Fatalf("got %q, want TestClan", result.Clan.Name)
	}
	if cache.setTag != "#AAA111" {
		t.Fatalf("expected cache set for #AAA111, got %q", cache.setTag)
	}
}

func TestClanServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakeCocapiClanClient{err: cocapi.ErrNotFound}
	cache := &fakeClanCache{clanHit: false}
	svc := NewClanService(api, cache)

	_, err := svc.FetchClan(context.Background(), "#AAA111")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeClanNotFound {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeClanNotFound)
	}
}

func TestClanServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewClanService(&fakeCocapiClanClient{}, &fakeClanCache{})
	_, err := svc.FetchClan(context.Background(), "bad tag!")
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeInvalidTag {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeInvalidTag)
	}
}
