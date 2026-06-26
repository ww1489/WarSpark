package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/repository"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type cocapiClanClient interface {
	GetClan(ctx context.Context, clanTag string) (cocapi.Clan, error)
	GetClanMembers(ctx context.Context, clanTag string, query cocapi.QueryGetClanMembers) (cocapi.ClanMemberListResponse, error)
}

type ClanRepository interface {
	SaveSnapshot(ctx context.Context, input repository.SaveClanSnapshotInput) error
}

type ClanCache interface {
	GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error)
	SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail) error
}

type ClanService struct {
	cocapi     cocapiClanClient
	cache      ClanCache
	repository ClanRepository
}

func NewClanService(cocapi cocapiClanClient, cache ClanCache, repository ClanRepository) *ClanService {
	return &ClanService{cocapi: cocapi, cache: cache, repository: repository}
}

func (s *ClanService) FetchClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, error) {
	normalizedTag, err := NormalizeClanTag(clanTag)
	if err != nil {
		return clandomain.ClanDetail{}, err
	}

	if s.cache != nil {
		detail, ok, err := s.cache.GetClan(ctx, normalizedTag)
		if err == nil && ok {
			return detail, nil
		}
	}

	clan, err := s.cocapi.GetClan(ctx, normalizedTag)
	if err != nil {
		return clandomain.ClanDetail{}, mapCocapiClanError(err)
	}
	membersResp, err := s.cocapi.GetClanMembers(ctx, normalizedTag, cocapi.QueryGetClanMembers{})
	if err != nil {
		return clandomain.ClanDetail{}, mapCocapiClanError(err)
	}
	detail := clandomain.ClanDetail{
		Clan:    toDomainClanOverview(clan),
		Members: toDomainClanMembers(membersResp.Items),
	}
	if s.cache != nil {
		_ = s.cache.SetClan(ctx, normalizedTag, detail)
	}
	if s.repository != nil {
		_ = s.repository.SaveSnapshot(ctx, repository.SaveClanSnapshotInput{
			ID:        uuid.NewString(),
			ClanTag:   normalizedTag,
			Detail:    detail,
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
	return detail, nil
}

func mapCocapiClanError(err error) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		return dmerrors.Wrap(dmerrors.ErrCodeClanNotFound, err.Error(), err)
	case errors.Is(err, cocapi.ErrInvalidTag):
		return dmerrors.Wrap(dmerrors.ErrCodeInvalidTag, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return dmerrors.Wrap(dmerrors.ErrCodeAPINotConfigured, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return dmerrors.Wrap(dmerrors.ErrCodeAPIAccessDenied, err.Error(), err)
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return dmerrors.Wrap(dmerrors.ErrCodeAPIResponseInvalid, err.Error(), err)
	}
	return dmerrors.Wrap(dmerrors.ErrCodeAPIRequestFailed, err.Error(), err)
}
