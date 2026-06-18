package service

import (
	"context"
	"testing"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

func TestLayoutServiceCreateLayoutDraftGeneratesID(t *testing.T) {
	repo := &fakeLayoutRepository{}
	service := NewLayoutService(repo)

	detail, err := service.CreateLayoutDraft(context.Background(), layoutdomain.CreateInput{
		Title:         "TH16 War Base",
		THLevel:       16,
		LayoutType:    "war",
		SourceType:    "manual_entry",
		ReviewStatus:  "pending_review",
		QualityStatus: "medium_confidence",
		Visibility:    "hidden",
	})
	if err != nil {
		t.Fatalf("CreateLayoutDraft returned error: %v", err)
	}
	if repo.createInput.ID == "" {
		t.Fatalf("expected generated layout id")
	}
	if detail.ID != repo.createInput.ID {
		t.Fatalf("expected detail id %s, got %s", repo.createInput.ID, detail.ID)
	}
}

func TestLayoutServiceAddLayoutLinkGeneratesID(t *testing.T) {
	repo := &fakeLayoutRepository{}
	service := NewLayoutService(repo)

	link, err := service.AddLayoutLink(context.Background(), "layout_123", layoutdomain.LinkInput{
		LinkType:   "official_open_layout",
		URL:        "https://link.clashofclans.com/example",
		LinkStatus: "unverified",
		SourceType: "manual_entry",
	})
	if err != nil {
		t.Fatalf("AddLayoutLink returned error: %v", err)
	}
	if repo.addLinkID == "" {
		t.Fatalf("expected generated link id")
	}
	if repo.addLinkLayoutID != "layout_123" {
		t.Fatalf("expected layout id layout_123, got %s", repo.addLinkLayoutID)
	}
	if link.ID != repo.addLinkID {
		t.Fatalf("expected link id %s, got %s", repo.addLinkID, link.ID)
	}
}

func TestLayoutServiceAddLayoutVideoMatchGeneratesID(t *testing.T) {
	repo := &fakeLayoutRepository{}
	service := NewLayoutService(repo)

	match, err := service.AddLayoutVideoMatch(context.Background(), "layout_123", layoutdomain.VideoMatchInput{
		YouTubeVideoID:   "abc123",
		VideoTitle:       "TH16 Attack Replay",
		TimestampSeconds: 245,
		MatchGroup:       "attack_video",
		MatchType:        "similar",
		SourceType:       "manual_entry",
		ReviewStatus:     "pending_review",
	})
	if err != nil {
		t.Fatalf("AddLayoutVideoMatch returned error: %v", err)
	}
	if repo.addVideoMatchID == "" {
		t.Fatalf("expected generated match id")
	}
	if repo.addVideoMatchLayoutID != "layout_123" {
		t.Fatalf("expected layout id layout_123, got %s", repo.addVideoMatchLayoutID)
	}
	if match.MatchID != repo.addVideoMatchID {
		t.Fatalf("expected match id %s, got %s", repo.addVideoMatchID, match.MatchID)
	}
}

func TestLayoutServiceListReviewQueuePassesFilterAndPagination(t *testing.T) {
	repo := &fakeLayoutRepository{reviewQueue: layoutdomain.ReviewQueueResult{Total: 7}}
	service := NewLayoutService(repo)

	result, err := service.ListReviewQueue(context.Background(), layoutdomain.ReviewQueueFilter{ResourceType: "layout"}, utils.Pagination{Page: 2, PageSize: 10, Offset: 10})
	if err != nil {
		t.Fatalf("ListReviewQueue returned error: %v", err)
	}
	if result.Total != 7 {
		t.Fatalf("expected total 7, got %d", result.Total)
	}
	if repo.reviewQueueFilter.ResourceType != "layout" || repo.reviewQueuePagination.Offset != 10 {
		t.Fatalf("unexpected call: filter=%#v pagination=%#v", repo.reviewQueueFilter, repo.reviewQueuePagination)
	}
}

func TestLayoutServiceListAuditLogsPassesPagination(t *testing.T) {
	repo := &fakeLayoutRepository{auditLogs: layoutdomain.AuditLogResult{Total: 3}}
	service := NewLayoutService(repo)

	result, err := service.ListAuditLogs(context.Background(), utils.Pagination{Page: 1, PageSize: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListAuditLogs returned error: %v", err)
	}
	if result.Total != 3 || repo.auditPagination.PageSize != 5 {
		t.Fatalf("unexpected result=%#v pagination=%#v", result, repo.auditPagination)
	}
}

func TestLayoutServiceUpdateReviewStatusPassesInput(t *testing.T) {
	repo := &fakeLayoutRepository{
		reviewResult: layoutdomain.ReviewUpdateResult{
			ResourceType:  "layout",
			ResourceID:    "layout_123",
			ReviewStatus:  "reviewed",
			QualityStatus: "high_confidence",
			Visibility:    "public",
		},
	}
	service := NewLayoutService(repo)

	result, err := service.UpdateReviewStatus(context.Background(), layoutdomain.ReviewUpdateInput{
		ResourceType:  "layout",
		ResourceID:    "layout_123",
		ReviewStatus:  "reviewed",
		QualityStatus: "high_confidence",
		Visibility:    "public",
		Note:          "source checked",
	})
	if err != nil {
		t.Fatalf("UpdateReviewStatus returned error: %v", err)
	}
	if repo.reviewInput.ResourceType != "layout" || repo.reviewInput.ResourceID != "layout_123" {
		t.Fatalf("unexpected review input: %#v", repo.reviewInput)
	}
	if repo.reviewInput.Note != "source checked" {
		t.Fatalf("expected note to be passed through, got %#v", repo.reviewInput)
	}
	if result.ReviewStatus != "reviewed" || result.Visibility != "public" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

type fakeLayoutRepository struct {
	createInput layoutdomain.CreateInput

	addLinkID       string
	addLinkLayoutID string
	addLinkInput    layoutdomain.LinkInput

	addVideoMatchID       string
	addVideoMatchLayoutID string
	addVideoMatchInput    layoutdomain.VideoMatchInput

	reviewQueueFilter     layoutdomain.ReviewQueueFilter
	reviewQueuePagination utils.Pagination
	reviewQueue           layoutdomain.ReviewQueueResult

	auditPagination utils.Pagination
	auditLogs       layoutdomain.AuditLogResult

	reviewInput  layoutdomain.ReviewUpdateInput
	reviewResult layoutdomain.ReviewUpdateResult
}

func (f *fakeLayoutRepository) List(context.Context, layoutdomain.ListFilter, utils.Pagination) (layoutdomain.ListResult, error) {
	return layoutdomain.ListResult{}, nil
}

func (f *fakeLayoutRepository) Get(_ context.Context, id string) (layoutdomain.Detail, error) {
	return layoutdomain.Detail{ID: id}, nil
}

func (f *fakeLayoutRepository) CreateDraft(_ context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error) {
	f.createInput = input
	return layoutdomain.Detail{
		ID:            input.ID,
		Title:         input.Title,
		THLevel:       input.THLevel,
		LayoutType:    input.LayoutType,
		StyleTags:     input.StyleTags,
		SourceType:    input.SourceType,
		SourceURL:     input.SourceURL,
		ReviewStatus:  input.ReviewStatus,
		QualityStatus: input.QualityStatus,
	}, nil
}

func (f *fakeLayoutRepository) AddLink(_ context.Context, id string, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error) {
	f.addLinkID = id
	f.addLinkLayoutID = layoutID
	f.addLinkInput = input
	return layoutdomain.Link{
		ID:         id,
		LinkType:   input.LinkType,
		URL:        input.URL,
		LinkStatus: input.LinkStatus,
		SourceType: input.SourceType,
		SourceURL:  input.SourceURL,
	}, nil
}

func (f *fakeLayoutRepository) UpdateLink(_ context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error) {
	return layoutdomain.Link{ID: linkID, LinkStatus: input.LinkStatus}, nil
}

func (f *fakeLayoutRepository) AddVideoMatch(_ context.Context, matchID string, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error) {
	f.addVideoMatchID = matchID
	f.addVideoMatchLayoutID = layoutID
	f.addVideoMatchInput = input
	return layoutdomain.VideoMatch{
		MatchID:          matchID,
		YouTubeVideoID:   input.YouTubeVideoID,
		VideoTitle:       input.VideoTitle,
		TimestampSeconds: input.TimestampSeconds,
		MatchGroup:       input.MatchGroup,
		MatchType:        input.MatchType,
		ReviewStatus:     input.ReviewStatus,
	}, nil
}

func (f *fakeLayoutRepository) ListReviewQueue(_ context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error) {
	f.reviewQueueFilter = filter
	f.reviewQueuePagination = pagination
	return f.reviewQueue, nil
}

func (f *fakeLayoutRepository) ListAuditLogs(_ context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error) {
	f.auditPagination = pagination
	return f.auditLogs, nil
}

func (f *fakeLayoutRepository) UpdateReviewStatus(_ context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	f.reviewInput = input
	return f.reviewResult, nil
}
