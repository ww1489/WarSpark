package service

import (
	"context"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type labelCache interface {
	GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error)
	SetClanLabels(ctx context.Context, resp label.LabelListResponse) error
	GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error)
	SetPlayerLabels(ctx context.Context, resp label.LabelListResponse) error
}

type cocapiLabelClient interface {
	GetClanLabels(ctx context.Context, query cocapi.QueryGetClanLabels) (cocapi.LabelListResponse, error)
	GetPlayerLabels(ctx context.Context, query cocapi.QueryGetPlayerLabels) (cocapi.LabelListResponse, error)
}

type LabelService struct {
	cocapi cocapiLabelClient
	cache  labelCache
}

func NewLabelService(cocapi cocapiLabelClient, cache labelCache) *LabelService {
	return &LabelService{cocapi: cocapi, cache: cache}
}

func (s *LabelService) GetClanLabels(ctx context.Context) (label.LabelListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetClanLabels(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetClanLabels(ctx, cocapi.QueryGetClanLabels{})
	if err != nil {
		return label.LabelListResponse{}, mapCocapiLabelError(err)
	}
	resp := label.LabelListResponse{Paging: toDomainLabelPaging(cocapiResp.Paging)}
	for _, lbl := range cocapiResp.Items {
		resp.Items = append(resp.Items, label.Label{ID: lbl.ID, Name: string(lbl.Name), IconURLs: lbl.IconURLs})
	}
	if s.cache != nil {
		_ = s.cache.SetClanLabels(ctx, resp)
	}
	return resp, nil
}

func (s *LabelService) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetPlayerLabels(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.cocapi.GetPlayerLabels(ctx, cocapi.QueryGetPlayerLabels{})
	if err != nil {
		return label.LabelListResponse{}, mapCocapiLabelError(err)
	}
	resp := label.LabelListResponse{Paging: toDomainLabelPaging(cocapiResp.Paging)}
	for _, lbl := range cocapiResp.Items {
		resp.Items = append(resp.Items, label.Label{ID: lbl.ID, Name: string(lbl.Name), IconURLs: lbl.IconURLs})
	}
	if s.cache != nil {
		_ = s.cache.SetPlayerLabels(ctx, resp)
	}
	return resp, nil
}

func toDomainLabelPaging(p cocapi.Paging) label.Paging {
	return label.Paging{Cursors: label.Cursors{After: p.Cursors.After, Before: p.Cursors.Before}}
}

func mapCocapiLabelError(err error) error {
	return mapCocapiRankingError(err, "")
}
