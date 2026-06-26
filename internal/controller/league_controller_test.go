package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/league"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type fakeLeagueService struct {
	leagues    league.LeagueListResponse
	l          league.League
	seasons    league.LeagueSeasonListResponse
	rankings   league.LeagueSeasonRankingListResponse
	tiers      league.LeagueTierListResponse
	tier       league.LeagueTier
	history    league.LeagueSeasonResultListResponse
	warLeagues league.WarLeagueListResponse
	warLeague  league.WarLeague
	err        error
}

func (f *fakeLeagueService) GetLeagues(ctx context.Context) (league.LeagueListResponse, error) { return f.leagues, f.err }
func (f *fakeLeagueService) GetLeague(ctx context.Context, id string) (league.League, error) { return f.l, f.err }
func (f *fakeLeagueService) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, error) { return f.seasons, f.err }
func (f *fakeLeagueService) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, error) { return f.rankings, f.err }
func (f *fakeLeagueService) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, error) { return f.tiers, f.err }
func (f *fakeLeagueService) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, error) { return f.tier, f.err }
func (f *fakeLeagueService) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, error) { return f.history, f.err }
func (f *fakeLeagueService) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, error) { return f.warLeagues, f.err }
func (f *fakeLeagueService) GetWarLeague(ctx context.Context, id string) (league.WarLeague, error) { return f.warLeague, f.err }

func TestLeagueControllerGetLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{leagues: league.LeagueListResponse{Items: []league.League{{Name: "Champion"}}}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues", ctrl.GetLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeague(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{l: league.League{Name: "Champion"}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id", ctrl.GetLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueSeasons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons", ctrl.GetLeagueSeasons)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueSeasonRankings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/rankings", ctrl.GetLeagueSeasonRankings)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/rankings", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueTiers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/tiers", ctrl.GetLeagueTiers)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/tiers", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueTier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/tiers/:tier", ctrl.GetLeagueTier)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/tiers/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetLeagueHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/tiers/:tier/history", ctrl.GetLeagueHistory)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/tiers/1/history?player_tag=%23P1ABC", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetWarLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{warLeagues: league.WarLeagueListResponse{Items: []league.WarLeague{{Name: "Master"}}}}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/war-leagues", ctrl.GetWarLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/war-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerGetWarLeague(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/war-leagues/:id", ctrl.GetWarLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/war-leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status = %d, want 200", rec.Code) }
}

func TestLeagueControllerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{err: wardomain.NewError(wardomain.ErrorLeagueNotFound, "not found")}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id", ctrl.GetLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/999", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound { t.Fatalf("status = %d, want 404", rec.Code) }
}

func TestLeagueControllerGetLeagueHistoryMissingTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLeagueService{}
	ctrl := NewLeagueController(svc)
	router := gin.New()
	router.GET("/leagues/:id/seasons/:season/tiers/:tier/history", ctrl.GetLeagueHistory)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/leagues/1/seasons/2026-01/tiers/1/history", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity { t.Fatalf("status = %d, want 422", rec.Code) }
}
