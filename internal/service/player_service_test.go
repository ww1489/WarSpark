package service

import (
	"context"
	"testing"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeCocapiPlayerClient struct {
	player    cocapi.Player
	battleLog cocapi.BattleLogEntryListResponse
	err       error
	gotTag    string
}

func (f *fakeCocapiPlayerClient) GetPlayer(ctx context.Context, playerTag string) (cocapi.Player, error) {
	f.gotTag = playerTag
	return f.player, f.err
}

func (f *fakeCocapiPlayerClient) GetBattleLog(ctx context.Context, playerTag string) (cocapi.BattleLogEntryListResponse, error) {
	f.gotTag = playerTag
	return f.battleLog, f.err
}

func TestPlayerServiceFetchReturnsCached(t *testing.T) {
	cached := clandomain.PlayerOverview{Name: "Cached"}
	cache := &fakeClanCache{playerOverview: cached, playerHit: true}
	svc := NewPlayerService(&fakeCocapiPlayerClient{}, cache, nil)

	result, err := svc.FetchPlayer(context.Background(), "#ABC123")
	if err != nil {
		t.Fatalf("FetchPlayer error: %v", err)
	}
	if result.Name != "Cached" {
		t.Fatalf("got %q, want Cached", result.Name)
	}
}

func TestPlayerServiceFetchCallsAPIOnMiss(t *testing.T) {
	api := &fakeCocapiPlayerClient{
		player: cocapi.Player{Name: "TestPlayer", Tag: "#ABC123", TownHallLevel: 15},
	}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, nil)

	result, err := svc.FetchPlayer(context.Background(), "#ABC123")
	if err != nil {
		t.Fatalf("FetchPlayer error: %v", err)
	}
	if result.Name != "TestPlayer" || result.TownHallLevel != 15 {
		t.Fatalf("got %+v", result)
	}
}

func TestPlayerServiceFetchBattleLog(t *testing.T) {
	api := &fakeCocapiPlayerClient{
		battleLog: cocapi.BattleLogEntryListResponse{
			Items: []cocapi.BattleLogEntry{{Stars: 3}},
		},
	}
	cache := &fakeClanCache{battleLogHit: false}
	svc := NewPlayerService(api, cache, nil)

	result, err := svc.FetchBattleLog(context.Background(), "#ABC123")
	if err != nil {
		t.Fatalf("FetchBattleLog error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Stars != 3 {
		t.Fatalf("got %+v", result)
	}
}

func TestPlayerServiceFetchPropagatesNotFound(t *testing.T) {
	api := &fakeCocapiPlayerClient{err: cocapi.ErrNotFound}
	cache := &fakeClanCache{playerHit: false}
	svc := NewPlayerService(api, cache, nil)

	_, err := svc.FetchPlayer(context.Background(), "#ABC123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodePlayerNotFound {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodePlayerNotFound)
	}
}

func TestPlayerServiceFetchRejectsInvalidTag(t *testing.T) {
	svc := NewPlayerService(&fakeCocapiPlayerClient{}, &fakeClanCache{}, nil)
	_, err := svc.FetchPlayer(context.Background(), "bad tag!")
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if dmerrors.Code(err) != dmerrors.ErrCodeInvalidTag {
		t.Fatalf("code = %q, want %q", dmerrors.Code(err), dmerrors.ErrCodeInvalidTag)
	}
}
