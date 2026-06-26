package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type fakeLabelService struct {
	clanLabels   label.LabelListResponse
	playerLabels label.LabelListResponse
	err          error
}

func (f *fakeLabelService) GetClanLabels(ctx context.Context) (label.LabelListResponse, error) {
	return f.clanLabels, f.err
}
func (f *fakeLabelService) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error) {
	return f.playerLabels, f.err
}

func TestLabelControllerGetClanLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLabelService{clanLabels: label.LabelListResponse{Items: []label.Label{{ID: 1, Name: "Clan War"}}}}
	ctrl := NewLabelController(svc)
	router := gin.New()
	router.GET("/clans/labels", ctrl.GetClanLabels)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/clans/labels", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestLabelControllerGetPlayerLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeLabelService{playerLabels: label.LabelListResponse{Items: []label.Label{{ID: 2, Name: "Active"}}}}
	ctrl := NewLabelController(svc)
	router := gin.New()
	router.GET("/players/labels", ctrl.GetPlayerLabels)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/labels", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
