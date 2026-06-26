package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeRankingService struct {
	locations      ranking.LocationListResponse
	clanRanking    ranking.ClanRankingListResponse
	playerRanking  ranking.PlayerRankingListResponse
	capitalRanking ranking.ClanCapitalRankingListResponse
	builderClan    ranking.ClanBuilderBaseRankingListResponse
	builderPlayer  ranking.PlayerBuilderBaseRankingListResponse
	err            error
}

func (f *fakeRankingService) GetLocations(ctx context.Context) (ranking.LocationListResponse, error) {
	return f.locations, f.err
}
func (f *fakeRankingService) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, error) {
	return f.clanRanking, f.err
}
func (f *fakeRankingService) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, error) {
	return f.playerRanking, f.err
}
func (f *fakeRankingService) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, error) {
	return f.capitalRanking, f.err
}
func (f *fakeRankingService) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, error) {
	return f.builderClan, f.err
}
func (f *fakeRankingService) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, error) {
	return f.builderPlayer, f.err
}

func TestRankingControllerGetLocations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{locations: ranking.LocationListResponse{Items: []ranking.Location{{ID: 1, Name: "Global"}}}}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations", ctrl.GetLocations)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("len = %d", len(items))
	}
}

func TestRankingControllerGetClanRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans", ctrl.GetClanRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/clans", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRankingControllerGetPlayerRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/players", ctrl.GetPlayerRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/players", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRankingControllerGetClanCapitalRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans-capital", ctrl.GetClanCapitalRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/clans-capital", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRankingControllerGetClanBuilderBaseRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans-builder-base", ctrl.GetClanBuilderBaseRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/clans-builder-base", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRankingControllerGetPlayerBuilderBaseRanking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/players-builder-base", ctrl.GetPlayerBuilderBaseRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/global/rankings/players-builder-base", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRankingControllerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeRankingService{err: wardomain.NewError(wardomain.ErrorLocationNotFound, "not found")}
	ctrl := NewRankingController(svc)
	router := gin.New()
	router.GET("/locations/:id/rankings/clans", ctrl.GetClanRanking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/999/rankings/clans", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
