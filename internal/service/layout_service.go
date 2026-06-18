package service

import (
	"context"

	"github.com/google/uuid"

	layoutdomain "github.com/ww1489/WarSpark/internal/domain/layout"
	"github.com/ww1489/WarSpark/internal/utils"
)

type LayoutRepository interface {
	List(ctx context.Context, filter layoutdomain.ListFilter, pagination utils.Pagination) (layoutdomain.ListResult, error)
	Get(ctx context.Context, id string) (layoutdomain.Detail, error)
	ListVideosByLayout(ctx context.Context, layoutID, matchGroup, matchType string) ([]layoutdomain.VideoMatch, error)
	CreateDraft(ctx context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error)
	AddLink(ctx context.Context, id string, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error)
	UpdateLink(ctx context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error)
	AddVideoMatch(ctx context.Context, matchID string, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error)
	ListReviewQueue(ctx context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error)
	ListAuditLogs(ctx context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error)
	UpdateReviewStatus(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error)
}

type LayoutService struct {
	repository LayoutRepository
}

func NewLayoutService(repository LayoutRepository) *LayoutService {
	return &LayoutService{repository: repository}
}

func (s *LayoutService) ListLayouts(ctx context.Context, filter layoutdomain.ListFilter, pagination utils.Pagination) (layoutdomain.ListResult, error) {
	return s.repository.List(ctx, filter, pagination)
}

func (s *LayoutService) GetLayout(ctx context.Context, id string) (layoutdomain.Detail, error) {
	return s.repository.Get(ctx, id)
}

func (s *LayoutService) ListVideos(ctx context.Context, layoutID, matchGroup, matchType string) ([]layoutdomain.VideoMatch, error) {
	return s.repository.ListVideosByLayout(ctx, layoutID, matchGroup, matchType)
}

func (s *LayoutService) CreateLayoutDraft(ctx context.Context, input layoutdomain.CreateInput) (layoutdomain.Detail, error) {
	if input.ID == "" {
		input.ID = uuid.NewString()
	}
	return s.repository.CreateDraft(ctx, input)
}

func (s *LayoutService) AddLayoutLink(ctx context.Context, layoutID string, input layoutdomain.LinkInput) (layoutdomain.Link, error) {
	return s.repository.AddLink(ctx, uuid.NewString(), layoutID, input)
}

func (s *LayoutService) UpdateLayoutLink(ctx context.Context, linkID string, input layoutdomain.LinkUpdateInput) (layoutdomain.Link, error) {
	return s.repository.UpdateLink(ctx, linkID, input)
}

func (s *LayoutService) AddLayoutVideoMatch(ctx context.Context, layoutID string, input layoutdomain.VideoMatchInput) (layoutdomain.VideoMatch, error) {
	return s.repository.AddVideoMatch(ctx, uuid.NewString(), layoutID, input)
}

func (s *LayoutService) ListReviewQueue(ctx context.Context, filter layoutdomain.ReviewQueueFilter, pagination utils.Pagination) (layoutdomain.ReviewQueueResult, error) {
	return s.repository.ListReviewQueue(ctx, filter, pagination)
}

func (s *LayoutService) ListAuditLogs(ctx context.Context, pagination utils.Pagination) (layoutdomain.AuditLogResult, error) {
	return s.repository.ListAuditLogs(ctx, pagination)
}

func (s *LayoutService) UpdateReviewStatus(ctx context.Context, input layoutdomain.ReviewUpdateInput) (layoutdomain.ReviewUpdateResult, error) {
	return s.repository.UpdateReviewStatus(ctx, input)
}
