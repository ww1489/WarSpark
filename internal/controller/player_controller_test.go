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

type fakePlayerService struct {
	player    clandomain.PlayerOverview
	battleLog clandomain.BattleLogSummary
	err       error
}

func (f *fakePlayerService) FetchPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, error) {
	if f.err != nil {
		return clandomain.PlayerOverview{}, f.err
	}
	return f.player, nil
}

func (f *fakePlayerService) FetchBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, error) {
	if f.err != nil {
		return clandomain.BattleLogSummary{}, f.err
	}
	return f.battleLog, nil
}

func TestPlayerControllerGetPlayerReturnsOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{player: clandomain.PlayerOverview{Name: "TestPlayer"}}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag", ctrl.GetPlayer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["name"] != "TestPlayer" {
		t.Fatalf("name = %v, want TestPlayer", data["name"])
	}
}

func TestPlayerControllerGetBattleLogReturnsItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{battleLog: clandomain.BattleLogSummary{Items: []clandomain.BattleLogEntry{{Stars: 3}}}}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag/battle-log", ctrl.GetBattleLog)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1/battle-log", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestPlayerControllerGetPlayerReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakePlayerService{err: dmerrors.New(dmerrors.ErrCodePlayerNotFound, "not found")}
	ctrl := NewPlayerController(svc)

	router := gin.New()
	router.GET("/players/:tag", ctrl.GetPlayer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/players/%23P1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
