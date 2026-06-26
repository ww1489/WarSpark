package service

import (
	"context"
	"testing"
	"time"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakePlayerAPIClient struct {
	player    clandomain.PlayerOverview
	battleLog clandomain.BattleLogSummary
	err       error
	gotTag    string
}

func (f *fakePlayerAPIClient) Player(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	f.gotTag = playerTag
	if f.err != nil {
		return clandomain.PlayerOverview{}, f.err
	}
	return f.player, nil
}

func (f *fakePlayerAPIClient) BattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	f.gotTag = playerTag
	if f.err != nil {
		return clandomain.BattleLogSummary{}, f.err
	}
	return f.battleLog, nil
}

func TestPlayerServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.PlayerOverview{Name: "Cached"}
	cache := &fakeClanCache{playerOverview: cached, playerHit: true}
	svc := NewPlayerService(&fakePlayerAPIClient{}, cache, 5*time.Minute)

	player, err := svc.FetchPlayer(context.Background(), "#PP1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if player.Name != "Cached" {
		t.Fatalf("got %q, want Cached", player.Name)
	}
}

func TestPlayerServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakePlayerAPIClient{player: clandomain.PlayerOverview{Name: "Fresh"}}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	player, err := svc.FetchPlayer(context.Background(), "#PP1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if player.Name != "Fresh" {
		t.Fatalf("got %q, want Fresh", player.Name)
	}
	if cache.setPlayerTag != "#PP1" {
		t.Fatalf("cache set tag %q, want #P1", cache.setPlayerTag)
	}
}

func TestPlayerServiceFetchBattleLog(t *testing.T) {
	api := &fakePlayerAPIClient{battleLog: clandomain.BattleLogSummary{Items: []clandomain.BattleLogEntry{{Stars: 3}}}}
	cache := &fakeClanCache{battleLogHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	log, err := svc.FetchBattleLog(context.Background(), "#PP1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(log.Items) != 1 || log.Items[0].Stars != 3 {
		t.Fatalf("items: %+v", log.Items)
	}
}

func TestPlayerServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakePlayerAPIClient{err: wardomain.NewError(wardomain.ErrorPlayerNotFound, "not found")}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, 5*time.Minute)

	_, err := svc.FetchPlayer(context.Background(), "#PP1")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorPlayerNotFound {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorPlayerNotFound)
	}
}

func TestPlayerServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewPlayerService(&fakePlayerAPIClient{}, &fakeClanCache{}, 5*time.Minute)
	_, err := svc.FetchPlayer(context.Background(), "!!!")
	if err == nil {
		t.Fatal("expected error")
	}
	if wardomain.ErrorCode(err) != wardomain.ErrorInvalidTag {
		t.Fatalf("code = %q, want %q", wardomain.ErrorCode(err), wardomain.ErrorInvalidTag)
	}
}
