package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ww1489/WarSpark/internal/domain/utility"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeUtilityCache struct {
	goldPass   utility.GoldPassSeason
	goldPassOK bool
}

func (f *fakeUtilityCache) GetGoldPass(ctx context.Context) (utility.GoldPassSeason, bool, error) {
	return f.goldPass, f.goldPassOK, nil
}

func (f *fakeUtilityCache) SetGoldPass(ctx context.Context, resp utility.GoldPassSeason) error {
	return nil
}

func fakeUtilityAPI() *fakeCocapiClient {
	return &fakeCocapiClient{
		err: nil,
	}
}

func TestGetCurrentGoldPassSeason(t *testing.T) {
	api := fakeUtilityAPI()
	cache := &fakeUtilityCache{}
	svc := NewUtilityService(api, cache)

	season, err := svc.GetCurrentGoldPassSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if season.EndTime != "" || season.StartTime != "" {
		t.Fatalf("expected empty season, got %+v", season)
	}
}

func TestGetCurrentGoldPassSeasonCached(t *testing.T) {
	cached := utility.GoldPassSeason{EndTime: "2026-07-01", StartTime: "2026-04-01"}
	api := fakeUtilityAPI()
	cache := &fakeUtilityCache{goldPass: cached, goldPassOK: true}
	svc := NewUtilityService(api, cache)

	season, err := svc.GetCurrentGoldPassSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if season.EndTime != "2026-07-01" {
		t.Fatalf("expected EndTime 2026-07-01, got %s", season.EndTime)
	}
}

func TestSearchClans(t *testing.T) {
	api := fakeUtilityAPI()
	svc := NewUtilityService(api, &fakeUtilityCache{})

	params := utility.ClanSearchParams{Name: "test"}
	resp, err := svc.SearchClans(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Fatalf("expected empty items, got %d", len(resp.Items))
	}
}

func TestSearchClansError(t *testing.T) {
	api := fakeUtilityAPI()
	api.err = cocapi.ErrNotFound
	svc := NewUtilityService(api, &fakeUtilityCache{})

	_, err := svc.SearchClans(context.Background(), utility.ClanSearchParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != wardomain.ErrorClanNotFound {
		t.Fatalf("expected code %s, got %s", wardomain.ErrorClanNotFound, warErr.Code)
	}
}

func TestGetLocation(t *testing.T) {
	api := fakeUtilityAPI()
	svc := NewUtilityService(api, &fakeUtilityCache{})

	loc, err := svc.GetLocation(context.Background(), "32000006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.ID != 0 {
		t.Fatalf("expected zero Location, got %+v", loc)
	}
}

func TestGetLocationError(t *testing.T) {
	api := fakeUtilityAPI()
	api.err = cocapi.ErrNotFound
	svc := NewUtilityService(api, &fakeUtilityCache{})

	_, err := svc.GetLocation(context.Background(), "999999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != wardomain.ErrorLocationNotFound {
		t.Fatalf("expected code %s, got %s", wardomain.ErrorLocationNotFound, warErr.Code)
	}
}

func TestVerifyPlayerToken(t *testing.T) {
	api := fakeUtilityAPI()
	svc := NewUtilityService(api, &fakeUtilityCache{})

	resp, err := svc.VerifyPlayerToken(context.Background(), "#ABC123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "" {
		t.Fatalf("expected empty response, got %+v", resp)
	}
}

func TestVerifyPlayerTokenError(t *testing.T) {
	api := fakeUtilityAPI()
	api.err = cocapi.ErrNotFound
	svc := NewUtilityService(api, &fakeUtilityCache{})

	_, err := svc.VerifyPlayerToken(context.Background(), "#INVALID", "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != wardomain.ErrorPlayerNotFound {
		t.Fatalf("expected code %s, got %s", wardomain.ErrorPlayerNotFound, warErr.Code)
	}
}

func TestGetPlayerLeagueGroupStub(t *testing.T) {
	api := fakeUtilityAPI()
	svc := NewUtilityService(api, &fakeUtilityCache{})

	_, err := svc.GetPlayerLeagueGroup(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != "not_implemented" {
		t.Fatalf("expected code not_implemented, got %s", warErr.Code)
	}
}

func TestUtilityMapErrorInvalidTag(t *testing.T) {
	err := mapCocapiUtilityError(cocapi.ErrInvalidTag, "")
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != wardomain.ErrorInvalidTag {
		t.Fatalf("expected code %s, got %s", wardomain.ErrorInvalidTag, warErr.Code)
	}
}

func TestUtilityMapErrorAPINotConfigured(t *testing.T) {
	err := mapCocapiUtilityError(cocapi.ErrAPINotConfigured, "")
	var warErr wardomain.Error
	if !errors.As(err, &warErr) {
		t.Fatalf("expected wardomain.Error, got %T", err)
	}
	if warErr.Code != wardomain.ErrorAPINotConfigured {
		t.Fatalf("expected code %s, got %s", wardomain.ErrorAPINotConfigured, warErr.Code)
	}
}
