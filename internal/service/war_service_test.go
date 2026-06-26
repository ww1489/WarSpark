package service

import (
	"context"
	"testing"
	"time"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func TestWarServiceFetchCurrentWarPersistsSnapshot(t *testing.T) {
	fetchedAt := time.Date(2026, 6, 6, 10, 0, 0, 0, time.UTC)
	client := &fakeWarAPIClient{
		war: wardomain.CurrentWar{
			State:    "inWar",
			TeamSize: 15,
			Clan: wardomain.WarClan{
				Tag:                   "#AAA111",
				Stars:                 20,
				DestructionPercentage: 88.5,
				Members: []wardomain.WarMember{
					{
						Tag:           "#P1",
						Name:          "Player One",
						TownHallLevel: 16,
						MapPosition:   1,
						Attacks:       []wardomain.WarAttack{{DefenderTag: "#O1", Stars: 3, DestructionPercentage: 100}},
					},
				},
			},
			Opponent: wardomain.WarClan{
				Tag:                   "#BBB222",
				Stars:                 18,
				DestructionPercentage: 84.2,
				Members: []wardomain.WarMember{
					{
						Tag:           "#O1",
						Name:          "Enemy One",
						TownHallLevel: 16,
						MapPosition:   1,
					},
				},
			},
		},
	}
	repository := &fakeWarRepository{
		snapshot: wardomain.Snapshot{ID: "war_123", ClanTag: "#AAA111", OpponentClanTag: "#BBB222", WarState: "inWar", TeamSize: 15, FetchedAt: fetchedAt},
	}
	service := NewWarService(client, repository)
	service.now = func() time.Time { return fetchedAt }

	result, err := service.FetchCurrentWar(context.Background(), "#aaa111")
	if err != nil {
		t.Fatalf("FetchCurrentWar returned error: %v", err)
	}

	if client.requestedTag != "#AAA111" {
		t.Fatalf("expected normalized tag #AAA111, got %q", client.requestedTag)
	}
	if repository.saved.ClanTag != "#AAA111" || repository.saved.OpponentClanTag != "#BBB222" {
		t.Fatalf("unexpected saved snapshot: %#v", repository.saved)
	}
	if len(repository.saved.Members) != 2 {
		t.Fatalf("expected clan and opponent members, got %#v", repository.saved.Members)
	}
	if len(repository.saved.Targets) != 1 || repository.saved.Targets[0].TargetName != "Enemy One" {
		t.Fatalf("expected one opponent target, got %#v", repository.saved.Targets)
	}
	if result.ID != "war_123" || result.WarState != "inWar" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWarServiceFetchCurrentWarRejectsInvalidTag(t *testing.T) {
	service := NewWarService(&fakeWarAPIClient{}, &fakeWarRepository{})

	_, err := service.FetchCurrentWar(context.Background(), "bad tag!")
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if got := wardomain.ErrorCode(err); got != "invalid_tag" {
		t.Fatalf("expected invalid_tag, got %q", got)
	}
}

func TestWarServiceListMembersPassesFilter(t *testing.T) {
	repository := &fakeWarRepository{
		membersResult: wardomain.MemberListResult{
			Items: []wardomain.Member{{ID: "member_123", Side: "opponent"}},
			Total: 1,
		},
	}
	service := NewWarService(&fakeWarAPIClient{}, repository)

	result, err := service.ListMembers(context.Background(), "war_123", "opponent", utils.Pagination{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListMembers returned error: %v", err)
	}

	if repository.memberSnapshotID != "war_123" || repository.memberSide != "opponent" {
		t.Fatalf("unexpected member query: snapshot=%s side=%s", repository.memberSnapshotID, repository.memberSide)
	}
	if result.Total != 1 || result.Items[0].ID != "member_123" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWarServiceFetchCurrentWarReturnsCachedSnapshot(t *testing.T) {
	cache := &fakeWarCache{
		getHit: true,
		snapshot: wardomain.Snapshot{
			ID:              "cached_war",
			ClanTag:         "#AAA111",
			OpponentClanTag: "#BBB222",
			WarState:        "inWar",
		},
	}
	client := &fakeWarAPIClient{}
	repository := &fakeWarRepository{}
	service := NewWarService(client, repository, WarServiceOptions{
		Cache: cache,
	})

	result, err := service.FetchCurrentWar(context.Background(), "#aaa111")
	if err != nil {
		t.Fatalf("FetchCurrentWar returned error: %v", err)
	}

	if result.ID != "cached_war" {
		t.Fatalf("expected cached snapshot, got %#v", result)
	}
	if client.requestedTag != "" {
		t.Fatalf("expected API client not called, got tag %q", client.requestedTag)
	}
	if repository.saved.ID != "" {
		t.Fatalf("expected repository not called, got %#v", repository.saved)
	}
	if cache.getTag != "#AAA111" {
		t.Fatalf("expected normalized cache tag, got %q", cache.getTag)
	}
}

func TestWarServiceFetchCurrentWarStoresSnapshotInCache(t *testing.T) {
	fetchedAt := time.Date(2026, 6, 6, 10, 0, 0, 0, time.UTC)
	cache := &fakeWarCache{}
	client := &fakeWarAPIClient{
		war: wardomain.CurrentWar{
			State:    "inWar",
			TeamSize: 15,
			Clan: wardomain.WarClan{
				Tag: "#AAA111",
			},
			Opponent: wardomain.WarClan{
				Tag: "#BBB222",
			},
		},
	}
	repository := &fakeWarRepository{
		snapshot: wardomain.Snapshot{ID: "war_123", ClanTag: "#AAA111", OpponentClanTag: "#BBB222", WarState: "inWar", FetchedAt: fetchedAt},
	}
	service := NewWarService(client, repository, WarServiceOptions{
		Cache: cache,
	})
	service.now = func() time.Time { return fetchedAt }

	result, err := service.FetchCurrentWar(context.Background(), "#aaa111")
	if err != nil {
		t.Fatalf("FetchCurrentWar returned error: %v", err)
	}

	if result.ID != "war_123" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if cache.setTag != "#AAA111" || cache.setSnapshot.ID != "war_123" {
		t.Fatalf("expected cached saved snapshot, got tag=%q snapshot=%#v", cache.setTag, cache.setSnapshot)
	}
}

type fakeWarAPIClient struct {
	requestedTag string
	war          wardomain.CurrentWar
	err          error
}

func (f *fakeWarAPIClient) CurrentWar(_ context.Context, clanTag string) (wardomain.CurrentWar, error) {
	f.requestedTag = clanTag
	return f.war, f.err
}

type fakeWarRepository struct {
	saved wardomain.SaveSnapshotInput

	snapshot wardomain.Snapshot

	memberSnapshotID string
	memberSide       string
	memberPagination utils.Pagination
	membersResult    wardomain.MemberListResult
}

func (f *fakeWarAPIClient) CWLGroup(_ context.Context, clanTag string) (wardomain.CWLGroup, error) {
	return wardomain.CWLGroup{}, nil
}
func (f *fakeWarRepository) SaveSnapshot(_ context.Context, input wardomain.SaveSnapshotInput) (wardomain.Snapshot, error) {
	f.saved = input
	return f.snapshot, nil
}

func (f *fakeWarRepository) ListMembers(_ context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error) {
	f.memberSnapshotID = snapshotID
	f.memberSide = side
	f.memberPagination = pagination
	return f.membersResult, nil
}

type fakeWarCache struct {
	getTag   string
	getHit   bool
	snapshot wardomain.Snapshot
	getErr   error

	setTag      string
	setSnapshot wardomain.Snapshot
	setErr      error

	warLogResp cocapi.ClanWarLogResponse
	cwlWarResp cocapi.ClanWar
}

func (f *fakeWarCache) GetCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, bool, error) {
	f.getTag = clanTag
	return f.snapshot, f.getHit, f.getErr
}

func (f *fakeWarCache) SetCurrentWar(ctx context.Context, clanTag string, snapshot wardomain.Snapshot) error {
	f.setTag = clanTag
	f.setSnapshot = snapshot
	return f.setErr
}

func (f *fakeWarCache) GetCWLGroup(_ context.Context, clanTag string) (wardomain.CWLGroup, bool, error) {
	return wardomain.CWLGroup{}, false, nil
}

func (f *fakeWarCache) SetCWLGroup(_ context.Context, clanTag string, group wardomain.CWLGroup) error {
	return nil
}

func TestWarServiceGetWarLog(t *testing.T) {
	cache := &fakeWarCache{
		getHit: true,
	}
	client := &fakeWarAPIClient{}
	repository := &fakeWarRepository{}
	service := NewWarService(client, repository, WarServiceOptions{
		Cache: cache,
	})
	result, err := service.GetWarLog(context.Background(), "#AAA111", 10, "", "")
	if err != nil {
		t.Fatalf("GetWarLog returned error: %v", err)
	}
	if cache.getTag != "#AAA111" {
		t.Fatalf("expected cache get for #AAA111, got %q", cache.getTag)
	}
	_ = result
}

func TestWarServiceGetWarLogInvalidTag(t *testing.T) {
	service := NewWarService(&fakeWarAPIClient{}, &fakeWarRepository{})
	_, err := service.GetWarLog(context.Background(), "bad tag!", 0, "", "")
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if got := wardomain.ErrorCode(err); got != "invalid_tag" {
		t.Fatalf("expected invalid_tag, got %q", got)
	}
}

func TestWarServiceGetCWLWar(t *testing.T) {
	cache := &fakeWarCache{
		getHit: true,
	}
	service := NewWarService(&fakeWarAPIClient{}, &fakeWarRepository{}, WarServiceOptions{
		Cache: cache,
	})
	result, err := service.GetCWLWar(context.Background(), "#WAR123")
	if err != nil {
		t.Fatalf("GetCWLWar returned error: %v", err)
	}
	_ = result
}

func TestWarServiceGetCWLWarInvalidTag(t *testing.T) {
	service := NewWarService(&fakeWarAPIClient{}, &fakeWarRepository{})
	_, err := service.GetCWLWar(context.Background(), "bad tag!")
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if got := wardomain.ErrorCode(err); got != "invalid_tag" {
		t.Fatalf("expected invalid_tag, got %q", got)
	}
}

// extend fakeWarCache for new methods

func (f *fakeWarCache) GetWarLog(_ context.Context, clanTag string) (cocapi.ClanWarLogResponse, bool, error) {
	f.getTag = clanTag
	return f.warLogResp, f.getHit, f.getErr
}

func (f *fakeWarCache) SetWarLog(_ context.Context, clanTag string, resp cocapi.ClanWarLogResponse) error {
	return nil
}

func (f *fakeWarCache) GetCWLWar(_ context.Context, warTag string) (cocapi.ClanWar, bool, error) {
	f.getTag = warTag
	return f.cwlWarResp, f.getHit, f.getErr
}

func (f *fakeWarCache) SetCWLWar(_ context.Context, warTag string, resp cocapi.ClanWar) error {
	return nil
}
