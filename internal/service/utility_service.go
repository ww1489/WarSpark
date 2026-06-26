package service

import (
	"context"
	"errors"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	"github.com/ww1489/WarSpark/internal/domain/utility"
	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type utilityCache interface {
	GetGoldPass(ctx context.Context) (utility.GoldPassSeason, bool, error)
	SetGoldPass(ctx context.Context, resp utility.GoldPassSeason) error
}

type cocapiUtilityClient interface {
	GetCurrentGoldPassSeason(ctx context.Context) (cocapi.GoldPassSeason, error)
	SearchClans(ctx context.Context, query cocapi.QuerySearchClans) (cocapi.ClanListResponse, error)
	GetLocation(ctx context.Context, locationId string) (cocapi.Location, error)
	VerifyToken(ctx context.Context, playerTag string, body cocapi.VerifyTokenRequest) (cocapi.VerifyTokenResponse, error)
}

type UtilityService struct {
	api   cocapiUtilityClient
	cache utilityCache
}

func NewUtilityService(api cocapiUtilityClient, cache utilityCache) *UtilityService {
	return &UtilityService{api: api, cache: cache}
}

func (s *UtilityService) GetCurrentGoldPassSeason(ctx context.Context) (utility.GoldPassSeason, error) {
	if s.cache != nil {
		resp, ok, err := s.cache.GetGoldPass(ctx)
		if err == nil && ok {
			return resp, nil
		}
	}
	cocapiResp, err := s.api.GetCurrentGoldPassSeason(ctx)
	if err != nil {
		return utility.GoldPassSeason{}, mapCocapiUtilityError(err, "")
	}
	resp := utility.GoldPassSeason{
		EndTime:   cocapiResp.EndTime,
		StartTime: cocapiResp.StartTime,
	}
	if s.cache != nil {
		_ = s.cache.SetGoldPass(ctx, resp)
	}
	return resp, nil
}

func (s *UtilityService) SearchClans(ctx context.Context, params utility.ClanSearchParams) (cocapi.ClanListResponse, error) {
	query := cocapi.QuerySearchClans{
		Name:          params.Name,
		WarFrequency:  params.WarFrequency,
		LocationID:    params.LocationID,
		MinMembers:    params.MinMembers,
		MaxMembers:    params.MaxMembers,
		MinClanPoints: params.MinClanPoints,
		MinClanLevel:  params.MinClanLevel,
		Limit:         params.Limit,
		After:         params.After,
		Before:        params.Before,
		LabelIds:      params.LabelIds,
	}
	resp, err := s.api.SearchClans(ctx, query)
	if err != nil {
		return cocapi.ClanListResponse{}, mapCocapiUtilityError(err, wardomain.ErrorClanNotFound)
	}
	return resp, nil
}

func (s *UtilityService) GetLocation(ctx context.Context, locationID string) (cocapi.Location, error) {
	resp, err := s.api.GetLocation(ctx, locationID)
	if err != nil {
		return cocapi.Location{}, mapCocapiUtilityError(err, wardomain.ErrorLocationNotFound)
	}
	return resp, nil
}

func (s *UtilityService) VerifyPlayerToken(ctx context.Context, playerTag, token string) (cocapi.VerifyTokenResponse, error) {
	body := cocapi.VerifyTokenRequest{Token: token}
	resp, err := s.api.VerifyToken(ctx, playerTag, body)
	if err != nil {
		return cocapi.VerifyTokenResponse{}, mapCocapiUtilityError(err, wardomain.ErrorPlayerNotFound)
	}
	return resp, nil
}

func (s *UtilityService) GetPlayerLeagueGroup(ctx context.Context) (utility.PlayerLeagueGroup, error) {
	return utility.PlayerLeagueGroup{}, wardomain.NewError("not_implemented", "league group not available, requires CWL round data")
}

func mapCocapiUtilityError(err error, notFoundCode string) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		if notFoundCode != "" {
			return wardomain.NewError(notFoundCode, err.Error())
		}
		return err
	case errors.Is(err, cocapi.ErrInvalidTag):
		return wardomain.NewError(wardomain.ErrorInvalidTag, err.Error())
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return wardomain.NewError(wardomain.ErrorAPINotConfigured, err.Error())
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return wardomain.NewError(wardomain.ErrorAPIAccessDenied, err.Error())
	case errors.Is(err, cocapi.ErrAPIRequestFailed):
		return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrRateLimited):
		return wardomain.NewError(wardomain.ErrorAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return wardomain.NewError(wardomain.ErrorAPIResponseInvalid, err.Error())
	default:
		return err
	}
}
