package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

func TestAdminLayoutControllerCreateDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		createdLayout: layoutdomain.Detail{ID: "layout_123", Title: "TH16 War Base", THLevel: 16},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.POST("/api/v1/admin/layouts", controller.CreateDraft)

	body := `{
		"title":"TH16 War Base",
		"th_level":16,
		"layout_type":"war",
		"style_tags":["ring"],
		"source_type":"manual_entry",
		"source_url":"https://example.com/source",
		"review_status":"pending_review",
		"quality_status":"medium_confidence",
		"visibility":"hidden"
	}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/layouts", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.createInput.Title != "TH16 War Base" || service.createInput.THLevel != 16 {
		t.Fatalf("unexpected create input: %#v", service.createInput)
	}
	if service.createInput.ReviewStatus != "pending_review" || service.createInput.Visibility != "hidden" {
		t.Fatalf("unexpected status defaults: %#v", service.createInput)
	}

	var response struct {
		Code int                 `json:"code"`
		Data layoutdomain.Detail `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != 0 || response.Data.ID != "layout_123" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestAdminLayoutControllerAddLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		addedLink: layoutdomain.Link{ID: "link_123", LinkType: "official_open_layout", URL: "https://link.clashofclans.com/example", LinkStatus: "unverified"},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.POST("/api/v1/admin/layouts/:layout_id/links", controller.AddLink)

	body := `{
		"link_type":"official_open_layout",
		"url":"https://link.clashofclans.com/example",
		"source_type":"manual_entry",
		"source_url":"https://example.com/source"
	}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/layouts/layout_123/links", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.addLinkLayoutID != "layout_123" {
		t.Fatalf("expected layout id layout_123, got %s", service.addLinkLayoutID)
	}
	if service.addLinkInput.LinkStatus != "unverified" {
		t.Fatalf("expected default unverified link status, got %#v", service.addLinkInput)
	}

	var response struct {
		Code int               `json:"code"`
		Data layoutdomain.Link `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID != "link_123" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestAdminLayoutControllerUpdateLinkStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		updatedLink: layoutdomain.Link{ID: "link_123", LinkType: "official_open_layout", URL: "https://link.clashofclans.com/example", LinkStatus: "broken"},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.PATCH("/api/v1/admin/layout-links/:link_id", controller.UpdateLink)

	body := `{"link_status":"broken","note":"cannot open"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/layout-links/link_123", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.updateLinkID != "link_123" || service.updateLinkInput.LinkStatus != "broken" {
		t.Fatalf("unexpected update request: id=%s input=%#v", service.updateLinkID, service.updateLinkInput)
	}

	var response struct {
		Code int               `json:"code"`
		Data layoutdomain.Link `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.LinkStatus != "broken" {
		t.Fatalf("unexpected response: %#v", response.Data)
	}
}

func TestAdminLayoutControllerUpdateReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		reviewResult: layoutdomain.ReviewUpdateResult{
			ResourceType:  "layout",
			ResourceID:    "layout_123",
			ReviewStatus:  "reviewed",
			QualityStatus: "high_confidence",
			Visibility:    "public",
		},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.PATCH("/api/v1/admin/review/:resource_type/:resource_id", controller.UpdateReviewStatus)

	body := `{"review_status":"reviewed","quality_status":"high_confidence","visibility":"public","note":"source checked"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/review/layout/layout_123", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.reviewInput.ResourceType != "layout" || service.reviewInput.ResourceID != "layout_123" {
		t.Fatalf("unexpected review target: %#v", service.reviewInput)
	}
	if service.reviewInput.ReviewStatus != "reviewed" || service.reviewInput.QualityStatus != "high_confidence" || service.reviewInput.Visibility != "public" {
		t.Fatalf("unexpected review input: %#v", service.reviewInput)
	}
	if service.reviewInput.Note != "source checked" {
		t.Fatalf("expected note to be passed through, got %#v", service.reviewInput)
	}

	var response struct {
		Code int                             `json:"code"`
		Data layoutdomain.ReviewUpdateResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ReviewStatus != "reviewed" || response.Data.Visibility != "public" {
		t.Fatalf("unexpected response: %#v", response.Data)
	}
}

func TestAdminLayoutControllerAddVideoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	score := 0.8
	service := &fakeAdminLayoutService{
		addedVideoMatch: layoutdomain.VideoMatch{
			MatchID:          "match_123",
			VideoID:          "video_123",
			YouTubeVideoID:   "abc123",
			VideoTitle:       "TH16 Attack Replay",
			ChannelName:      "Example Channel",
			TimestampSeconds: 245,
			YouTubeURL:       "https://www.youtube.com/watch?v=abc123&t=245s",
			MatchGroup:       "attack_video",
			MatchType:        "similar",
			ConfidenceScore:  &score,
			ReviewStatus:     "pending_review",
		},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.POST("/api/v1/admin/layouts/:layout_id/video-matches", controller.AddVideoMatch)

	body := `{
		"youtube_video_id":"abc123",
		"video_title":"TH16 Attack Replay",
		"channel_name":"Example Channel",
		"timestamp_seconds":245,
		"match_group":"attack_video",
		"match_type":"similar",
		"stars":3,
		"destruction_percent":100,
		"confidence_score":0.8,
		"source_type":"manual_entry",
		"source_url":"https://example.com/source"
	}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/layouts/layout_123/video-matches", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.addVideoMatchLayoutID != "layout_123" {
		t.Fatalf("expected layout id layout_123, got %s", service.addVideoMatchLayoutID)
	}
	if service.addVideoMatchInput.ReviewStatus != "pending_review" {
		t.Fatalf("expected default pending_review, got %#v", service.addVideoMatchInput)
	}

	var response struct {
		Code int                     `json:"code"`
		Data layoutdomain.VideoMatch `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.MatchID != "match_123" || response.Data.YouTubeVideoID != "abc123" {
		t.Fatalf("unexpected response: %#v", response.Data)
	}
}

func TestAdminLayoutControllerListReviewQueue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		reviewQueue: layoutdomain.ReviewQueueResult{
			Items: []layoutdomain.ReviewQueueItem{
				{
					ResourceType:  "layout",
					ResourceID:    "layout_123",
					Title:         "TH16 War Base",
					ReviewStatus:  "pending_review",
					QualityStatus: "medium_confidence",
				},
			},
			Total: 1,
		},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.GET("/api/v1/admin/review-queue", controller.ListReviewQueue)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/review-queue?resource_type=layout&review_status=pending_review&page=2&page_size=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.reviewQueueFilter.ResourceType != "layout" || service.reviewQueueFilter.ReviewStatus != "pending_review" {
		t.Fatalf("unexpected filter: %#v", service.reviewQueueFilter)
	}
	if service.reviewQueuePagination.Page != 2 || service.reviewQueuePagination.PageSize != 10 {
		t.Fatalf("unexpected pagination: %#v", service.reviewQueuePagination)
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			Items      []layoutdomain.ReviewQueueItem `json:"items"`
			Pagination struct {
				Total int `json:"total"`
			} `json:"pagination"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.Pagination.Total != 1 || response.Data.Items[0].ResourceID != "layout_123" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestAdminLayoutControllerListAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeAdminLayoutService{
		auditLogs: layoutdomain.AuditLogResult{
			Items: []layoutdomain.AuditLog{
				{ID: "audit_123", ResourceType: "layout", ResourceID: "layout_123", Action: "create_draft"},
			},
			Total: 1,
		},
	}
	router := gin.New()
	controller := NewAdminLayoutController(service)
	router.GET("/api/v1/admin/audit-logs", controller.ListAuditLogs)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?page=1&page_size=5", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if service.auditPagination.PageSize != 5 {
		t.Fatalf("unexpected pagination: %#v", service.auditPagination)
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			Items []layoutdomain.AuditLog `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.Items[0].Action != "create_draft" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

type fakeAdminLayoutService struct {
	createInput   layoutdomain.CreateInput
	createdLayout layoutdomain.Detail

	addLinkLayoutID string
	addLinkInput    layoutdomain.LinkInput
	addedLink       layoutdomain.Link

	updateLinkID    string
	updateLinkInput layoutdomain.LinkUpdateInput
	updatedLink     layoutdomain.Link

	reviewInput  layoutdomain.ReviewUpdateInput
	reviewResult layoutdomain.ReviewUpdateResult

	addVideoMatchLayoutID string
	addVideoMatchInput    layoutdomain.VideoMatchInput
	addedVideoMatch       layoutdomain.VideoMatch

	reviewQueueFilter     layoutdomain.ReviewQueueFilter
	reviewQueuePagination utils.Pagination
	reviewQueue           layoutdomain.ReviewQueueResult

	auditPagination utils.Pagination
	auditLogs       layoutdomain.AuditLogResult
}

func (f *fakeAdminLayoutService) CreateLayoutDraft(_ context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error) {
	f.createInput = input
	return f.createdLayout, nil
}

func (f *fakeAdminLayoutService) AddLayoutLink(_ context.Context, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error) {
	f.addLinkLayoutID = layoutID
	f.addLinkInput = input
	return f.addedLink, nil
}

func (f *fakeAdminLayoutService) UpdateLayoutLink(_ context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error) {
	f.updateLinkID = linkID
	f.updateLinkInput = input
	return f.updatedLink, nil
}

func (f *fakeAdminLayoutService) UpdateReviewStatus(_ context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	f.reviewInput = input
	return f.reviewResult, nil
}

func (f *fakeAdminLayoutService) AddLayoutVideoMatch(_ context.Context, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error) {
	f.addVideoMatchLayoutID = layoutID
	f.addVideoMatchInput = input
	return f.addedVideoMatch, nil
}

func (f *fakeAdminLayoutService) ListReviewQueue(_ context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error) {
	f.reviewQueueFilter = filter
	f.reviewQueuePagination = pagination
	return f.reviewQueue, nil
}

func (f *fakeAdminLayoutService) ListAuditLogs(_ context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error) {
	f.auditPagination = pagination
	return f.auditLogs, nil
}
