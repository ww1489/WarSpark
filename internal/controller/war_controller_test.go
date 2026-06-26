package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func TestWarControllerGetCurrentWar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		currentWar: wardomain.Snapshot{
			ID:                  "war_123",
			ClanTag:             "#AAA111",
			OpponentClanTag:     "#BBB222",
			WarState:            "inWar",
			TeamSize:            15,
			ClanStars:           ptrInt(20),
			OpponentStars:       ptrInt(18),
			ClanDestruction:     ptrFloat64(88.5),
			OpponentDestruction: ptrFloat64(84.2),
			FetchedAt:           time.Date(2026, 6, 6, 10, 0, 0, 0, time.UTC),
		},
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/war/current", controller.GetCurrent)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/war/current?clan_tag=%23aaa111", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.currentWarTag != "#aaa111" {
		t.Fatalf("expected raw clan tag passed to service, got %q", service.currentWarTag)
	}

	var response struct {
		Code int                `json:"code"`
		Data wardomain.Snapshot `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID != "war_123" || response.Data.OpponentClanTag != "#BBB222" {
		t.Fatalf("unexpected response: %#v", response.Data)
	}
}

func TestWarControllerGetCurrentWarRequiresClanTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewWarController(&fakeWarService{})

	router := gin.New()
	router.GET("/api/v1/war/current", controller.GetCurrent)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/war/current", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestWarControllerListMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		members: wardomain.MemberListResult{
			Items: []wardomain.Member{{ID: "member_123", Side: "opponent"}},
			Total: 1,
		},
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/war/snapshots/:war_snapshot_id/members", controller.ListMembers)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/war/snapshots/war_123/members?side=opponent&page=2&page_size=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.memberSnapshotID != "war_123" || service.memberSide != "opponent" {
		t.Fatalf("unexpected query: snapshot=%s side=%s", service.memberSnapshotID, service.memberSide)
	}
	if service.memberPagination.Page != 2 || service.memberPagination.PageSize != 10 {
		t.Fatalf("unexpected pagination: %#v", service.memberPagination)
	}
}

func (f *fakeWarService) FetchCWLGroup(_ context.Context, clanTag string) (wardomain.CWLGroup, error) {
	return wardomain.CWLGroup{}, nil
}

type fakeWarService struct {
	currentWarTag string
	currentWar    wardomain.Snapshot

	memberSnapshotID string
	memberSide       string
	memberPagination utils.Pagination
	members          wardomain.MemberListResult

	warLogTag    string
	warLogLimit  int
	warLogAfter  string
	warLogBefore string
	warLogResp   cocapi.ClanWarLogResponse
	warLogErr    error

	cwlWarTag  string
	cwlWarResp cocapi.ClanWar
	cwlWarErr  error
}

func (f *fakeWarService) FetchCurrentWar(_ context.Context, clanTag string) (wardomain.Snapshot, error) {
	f.currentWarTag = clanTag
	return f.currentWar, nil
}

func (f *fakeWarService) ListMembers(_ context.Context, snapshotID string, side string, pagination utils.Pagination) (wardomain.MemberListResult, error) {
	f.memberSnapshotID = snapshotID
	f.memberSide = side
	f.memberPagination = pagination
	return f.members, nil
}

func (f *fakeWarService) GetWarLog(_ context.Context, clanTag string, limit int, after, before string) (cocapi.ClanWarLogResponse, error) {
	f.warLogTag = clanTag
	f.warLogLimit = limit
	f.warLogAfter = after
	f.warLogBefore = before
	return f.warLogResp, f.warLogErr
}

func (f *fakeWarService) GetCWLWar(_ context.Context, warTag string) (cocapi.ClanWar, error) {
	f.cwlWarTag = warTag
	return f.cwlWarResp, f.cwlWarErr
}

func TestWarControllerGetWarLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		warLogResp: cocapi.ClanWarLogResponse{
			Items: []cocapi.ClanWarLogEntry{
				{
					Result:   "win",
					TeamSize: 15,
				},
			},
		},
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/clans/:tag/war-log", controller.GetWarLog)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/clans/%23aaa111/war-log?limit=5", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.warLogTag != "#aaa111" {
		t.Fatalf("expected war log tag #aaa111, got %q", service.warLogTag)
	}
	if service.warLogLimit != 5 {
		t.Fatalf("expected limit 5, got %d", service.warLogLimit)
	}
}

func TestWarControllerGetWarLogServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		warLogErr: wardomain.NewError(wardomain.ErrorWarNotFound, "not found"),
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/clans/:tag/war-log", controller.GetWarLog)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/clans/%23aaa111/war-log", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestWarControllerGetCWLWar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		cwlWarResp: cocapi.ClanWar{
			State: "inWar",
		},
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/war/cwl/wars/:war_tag", controller.GetCWLWar)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/war/cwl/wars/%23war123", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.cwlWarTag != "#war123" {
		t.Fatalf("expected cwl war tag #war123, got %q", service.cwlWarTag)
	}
}

func TestWarControllerGetCWLWarNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeWarService{
		cwlWarErr: wardomain.NewError(wardomain.ErrorWarNotFound, "not found"),
	}
	controller := NewWarController(service)

	router := gin.New()
	router.GET("/api/v1/war/cwl/wars/:war_tag", controller.GetCWLWar)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/war/cwl/wars/%23war123", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func ptrInt(value int) *int {
	return &value
}

func ptrFloat64(value float64) *float64 {
	return &value
}
