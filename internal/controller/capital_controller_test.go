package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/capital"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
)

type fakeCapitalService struct {
	raidSeasons    capital.CapitalRaidSeasonListResponse
	capLeagues     capital.CapitalLeagueListResponse
	capLeague      capital.CapitalLeague
	builderLeagues capital.BuilderBaseLeagueListResponse
	builderLeague  capital.BuilderBaseLeague
	err            error
}

func (f *fakeCapitalService) GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, error) {
	return f.raidSeasons, f.err
}
func (f *fakeCapitalService) GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, error) {
	return f.capLeagues, f.err
}
func (f *fakeCapitalService) GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, error) {
	return f.capLeague, f.err
}
func (f *fakeCapitalService) GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, error) {
	return f.builderLeagues, f.err
}
func (f *fakeCapitalService) GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, error) {
	return f.builderLeague, f.err
}

func TestCapitalControllerGetCapitalRaidSeasons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{
		raidSeasons: capital.CapitalRaidSeasonListResponse{
			Items: []capital.CapitalRaidSeason{{State: "inactive"}},
		},
	}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/clans/:tag/capital-raid-seasons", ctrl.GetCapitalRaidSeasons)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%232PP/capital-raid-seasons", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCapitalControllerGetCapitalRaidSeasonsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{err: dmerrors.New(dmerrors.ErrCodeInvalidTag, "bad tag")}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/clans/:tag/capital-raid-seasons", ctrl.GetCapitalRaidSeasons)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%232PP/capital-raid-seasons", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestCapitalControllerGetCapitalLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{
		capLeagues: capital.CapitalLeagueListResponse{
			Items: []capital.CapitalLeague{{ID: 1, Name: "Bronze"}},
		},
	}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/capital-leagues", ctrl.GetCapitalLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capital-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCapitalControllerGetCapitalLeaguesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{err: dmerrors.New(dmerrors.ErrCodeAPINotConfigured, "not configured")}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/capital-leagues", ctrl.GetCapitalLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capital-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestCapitalControllerGetCapitalLeague(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{capLeague: capital.CapitalLeague{ID: 1, Name: "Bronze III"}}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/capital-leagues/:id", ctrl.GetCapitalLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capital-leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCapitalControllerGetCapitalLeagueError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{err: dmerrors.New(dmerrors.ErrCodeAPIAccessDenied, "denied")}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/capital-leagues/:id", ctrl.GetCapitalLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/capital-leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCapitalControllerGetBuilderBaseLeagues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{
		builderLeagues: capital.BuilderBaseLeagueListResponse{
			Items: []capital.BuilderBaseLeague{{ID: 1, Name: "Wood"}},
		},
	}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/builder-base-leagues", ctrl.GetBuilderBaseLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/builder-base-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCapitalControllerGetBuilderBaseLeaguesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{err: dmerrors.New(dmerrors.ErrCodeAPIRequestFailed, "bad gateway")}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/builder-base-leagues", ctrl.GetBuilderBaseLeagues)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/builder-base-leagues", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestCapitalControllerGetBuilderBaseLeague(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{builderLeague: capital.BuilderBaseLeague{ID: 1, Name: "Wood III"}}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/builder-base-leagues/:id", ctrl.GetBuilderBaseLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/builder-base-leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCapitalControllerGetBuilderBaseLeagueError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeCapitalService{err: dmerrors.New(dmerrors.ErrCodeWarNotFound, "not found")}
	ctrl := NewCapitalController(svc)
	router := gin.New()
	router.GET("/builder-base-leagues/:id", ctrl.GetBuilderBaseLeague)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/builder-base-leagues/1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
