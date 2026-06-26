package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
)

type fakeClanService struct {
	detail clandomain.ClanDetail
	err    error
}

func (f *fakeClanService) FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	if f.err != nil {
		return clandomain.ClanDetail{}, f.err
	}
	return f.detail, nil
}

func TestClanControllerGetClanReturnsDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeClanService{detail: clandomain.ClanDetail{Clan: clandomain.ClanOverview{Name: "Test Clan"}}}
	ctrl := NewClanController(svc)

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%23AAA", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	clan := data["clan"].(map[string]any)
	if clan["name"] != "Test Clan" {
		t.Fatalf("name = %v, want Test Clan", clan["name"])
	}
}

func TestClanControllerGetClanReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeClanService{err: dmerrors.New(dmerrors.ErrCodeClanNotFound, "not found")}
	ctrl := NewClanController(svc)

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/%23AAA", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestClanControllerGetClanRejectsMissingTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewClanController(&fakeClanService{})

	router := gin.New()
	router.GET("/clans/:tag", ctrl.GetClan)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d (gin router no match expected)", rec.Code)
	}
}
