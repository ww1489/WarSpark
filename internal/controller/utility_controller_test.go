package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/utility"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	"github.com/ww1489/WarSpark/internal/utils"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeUtilityService struct {
	goldPass   utility.GoldPassSeason
	clans      cocapi.ClanListResponse
	location   cocapi.Location
	verifyResp cocapi.VerifyTokenResponse
	leagueGrp  utility.PlayerLeagueGroup
	err        error
}

func (f *fakeUtilityService) GetCurrentGoldPassSeason(ctx context.Context) (utility.GoldPassSeason, error) {
	return f.goldPass, f.err
}

func (f *fakeUtilityService) SearchClans(ctx context.Context, params utility.ClanSearchParams) (cocapi.ClanListResponse, error) {
	return f.clans, f.err
}

func (f *fakeUtilityService) GetLocation(ctx context.Context, locationID string) (cocapi.Location, error) {
	return f.location, f.err
}

func (f *fakeUtilityService) VerifyPlayerToken(ctx context.Context, playerTag, token string) (cocapi.VerifyTokenResponse, error) {
	return f.verifyResp, f.err
}

func (f *fakeUtilityService) GetPlayerLeagueGroup(ctx context.Context) (utility.PlayerLeagueGroup, error) {
	return f.leagueGrp, f.err
}

func TestUtilityGetCurrentGoldPass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{goldPass: utility.GoldPassSeason{EndTime: "2026-07-01", StartTime: "2026-04-01"}}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.GET("/goldpass", ctrl.GetCurrentGoldPass)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/goldpass", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp utils.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("code = %d, want 0", resp.Code)
	}
}

func TestUtilitySearchClans(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{clans: cocapi.ClanListResponse{Items: []cocapi.Clan{{Name: "TestClan", Tag: "#ABC"}}}}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.GET("/clans", ctrl.SearchClans)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans?name=test", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestUtilityGetLocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{location: cocapi.Location{ID: 32000006, Name: "Global"}}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.GET("/locations/:id", ctrl.GetLocation)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/32000006", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestUtilityVerifyPlayerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{verifyResp: cocapi.VerifyTokenResponse{Status: "ok", Tag: "#ABC", Token: "test-token"}}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.POST("/players/:tag/verifytoken", ctrl.VerifyPlayerToken)
	rec := httptest.NewRecorder()
	body := `{"token":"test-token"}`
	req := httptest.NewRequest(http.MethodPost, "/players/%23ABC/verifytoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestUtilityVerifyPlayerTokenValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.POST("/players/:tag/verifytoken", ctrl.VerifyPlayerToken)
	rec := httptest.NewRecorder()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/players/%23ABC/verifytoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 422 or 400", rec.Code)
	}
}

func TestUtilityGetPlayerLeagueGroupStub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{err: wardomain.NewError("not_implemented", "league group not available, requires CWL round data")}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.GET("/players/leaguegroup", ctrl.GetPlayerLeagueGroup)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/leaguegroup", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUtilityGetLocationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeUtilityService{location: cocapi.Location{}, err: wardomain.NewError(wardomain.ErrorLocationNotFound, "location not found")}
	ctrl := NewUtilityController(svc)
	router := gin.New()
	router.GET("/locations/:id", ctrl.GetLocation)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/locations/999", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
