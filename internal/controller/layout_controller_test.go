package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

func TestLayoutControllerListLayoutsUsesFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeLayoutService{
		listResult: layoutdomain.ListResult{
			Items: []layoutdomain.Card{
				{
					ID:              "layout_123",
					Title:           "TH16 War Base",
					THLevel:         16,
					LayoutType:      "war",
					StyleTags:       []string{"ring", "box"},
					PrimaryImageURL: "https://storage.example/layouts/layout_123.png",
					SourceType:      "public_page",
					ReviewStatus:    "reviewed",
					QualityStatus:   "high_confidence",
					LinkStatus:      "active",
					UpdatedAt:       time.Date(2026, 6, 5, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
				},
			},
			Total: 42,
		},
	}
	router := gin.New()
	controller := NewLayoutController(service)
	router.GET("/api/v1/layouts", controller.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/layouts?th_level=16&layout_type=war&style_tag=ring&source_type=public_page&review_status=reviewed&quality_status=high_confidence&link_status=active&page=2&page_size=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.filter.THLevel == nil || *service.filter.THLevel != 16 {
		t.Fatalf("expected TH filter 16, got %#v", service.filter.THLevel)
	}
	if service.filter.LayoutType != "war" || service.filter.StyleTag != "ring" || service.filter.SourceType != "public_page" {
		t.Fatalf("unexpected filter: %#v", service.filter)
	}
	if service.pagination.Page != 2 || service.pagination.PageSize != 10 || service.pagination.Offset != 10 {
		t.Fatalf("unexpected pagination: %#v", service.pagination)
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			Items      []layoutdomain.Card `json:"items"`
			Pagination struct {
				Page     int `json:"page"`
				PageSize int `json:"page_size"`
				Total    int `json:"total"`
			} `json:"pagination"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("expected code 0, got %d", response.Code)
	}
	if len(response.Data.Items) != 1 || response.Data.Items[0].ID != "layout_123" {
		t.Fatalf("unexpected items: %#v", response.Data.Items)
	}
	if response.Data.Pagination.Total != 42 {
		t.Fatalf("expected total 42, got %d", response.Data.Pagination.Total)
	}
}

func TestLayoutControllerGetLayoutReturnsDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeLayoutService{
		detail: layoutdomain.Detail{
			ID:            "layout_123",
			Title:         "TH16 War Base",
			THLevel:       16,
			LayoutType:    "war",
			StyleTags:     []string{"ring"},
			SourceType:    "public_page",
			ReviewStatus:  "reviewed",
			QualityStatus: "high_confidence",
			Images: []layoutdomain.Image{
				{ID: "img_123", ImageURL: "https://storage.example/layouts/layout_123.png", ImageRole: "primary"},
			},
			Links: []layoutdomain.Link{
				{ID: "link_123", LinkType: "official_open_layout", URL: "https://link.clashofclans.com/example", LinkStatus: "active"},
			},
		},
	}
	router := gin.New()
	controller := NewLayoutController(service)
	router.GET("/api/v1/layouts/:layout_id", controller.Get)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/layouts/layout_123", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.detailID != "layout_123" {
		t.Fatalf("expected detail id layout_123, got %s", service.detailID)
	}

	var response struct {
		Code int                 `json:"code"`
		Data layoutdomain.Detail `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID != "layout_123" || len(response.Data.Images) != 1 || len(response.Data.Links) != 1 {
		t.Fatalf("unexpected detail: %#v", response.Data)
	}
}

func TestLayoutControllerGetLayoutReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeLayoutService{detailErr: layoutdomain.ErrLayoutNotFound}
	router := gin.New()
	controller := NewLayoutController(service)
	router.GET("/api/v1/layouts/:layout_id", controller.Get)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/layouts/missing", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, recorder.Code, recorder.Body.String())
	}
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != int(utils.ErrNotFound) || response.Message != "layout not found" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func (f *fakeLayoutService) ListVideos(_ context.Context, layoutID, matchGroup, matchType string) ([]layoutdomain.VideoMatch, error) {
	return nil, nil
}

type fakeLayoutService struct {
	filter     layoutdomain.ListFilter
	pagination utils.Pagination
	listResult layoutdomain.ListResult
	listErr    error

	detailID  string
	detail    layoutdomain.Detail
	detailErr error
}

func (f *fakeLayoutService) ListLayouts(_ context.Context, filter layoutdomain.ListFilter, pagination utils.Pagination) (layoutdomain.ListResult, error) {
	f.filter = filter
	f.pagination = pagination
	return f.listResult, f.listErr
}

func (f *fakeLayoutService) GetLayout(_ context.Context, id string) (layoutdomain.Detail, error) {
	f.detailID = id
	if f.detailErr != nil {
		return layoutdomain.Detail{}, f.detailErr
	}
	if f.detail.ID == "" {
		return layoutdomain.Detail{}, errors.New("missing fixture detail")
	}
	return f.detail, nil
}
