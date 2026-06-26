package service

import (
	"context"
	"errors"
	"strconv"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"

	dmerrors "github.com/ww1489/WarSpark/internal/domain/errors"
	"github.com/ww1489/WarSpark/internal/domain/utility"
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
	GetPlayer(ctx context.Context, playerTag string) (cocapi.Player, error)
	GetLeagueGroup(ctx context.Context, leagueGroupTag string, leagueSeasonId string, query cocapi.QueryGetLeagueGroup) (cocapi.LeagueGroup, error)
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
		return cocapi.ClanListResponse{}, mapCocapiUtilityError(err, dmerrors.ErrCodeClanNotFound)
	}
	return resp, nil
}

func (s *UtilityService) GetLocation(ctx context.Context, locationID string) (cocapi.Location, error) {
	resp, err := s.api.GetLocation(ctx, locationID)
	if err != nil {
		return cocapi.Location{}, mapCocapiUtilityError(err, dmerrors.ErrCodeLocationNotFound)
	}
	return resp, nil
}

func (s *UtilityService) VerifyPlayerToken(ctx context.Context, playerTag, token string) (cocapi.VerifyTokenResponse, error) {
	body := cocapi.VerifyTokenRequest{Token: token}
	resp, err := s.api.VerifyToken(ctx, playerTag, body)
	if err != nil {
		return cocapi.VerifyTokenResponse{}, mapCocapiUtilityError(err, dmerrors.ErrCodePlayerNotFound)
	}
	return resp, nil
}

func (s *UtilityService) GetPlayerLeagueGroup(ctx context.Context, playerTag string) (utility.PlayerLeagueGroup, error) {
	if playerTag == "" {
		return utility.PlayerLeagueGroup{}, dmerrors.New(dmerrors.ErrCodeInvalidTag, "player tag is required")
	}
	player, err := s.api.GetPlayer(ctx, playerTag)
	if err != nil {
		return utility.PlayerLeagueGroup{}, mapCocapiUtilityError(err, dmerrors.ErrCodePlayerNotFound)
	}
	if player.CurrentLeagueGroupTag == "" {
		return utility.PlayerLeagueGroup{}, nil
	}
	lg, err := s.api.GetLeagueGroup(ctx, player.CurrentLeagueGroupTag, strconv.FormatInt(int64(player.CurrentLeagueSeasonID), 10), cocapi.QueryGetLeagueGroup{})
	if err != nil {
		return utility.PlayerLeagueGroup{}, mapCocapiUtilityError(err, "")
	}
	return toDomainLeagueGroup(lg), nil
}

func toDomainLeagueGroup(lg cocapi.LeagueGroup) utility.PlayerLeagueGroup {
	result := utility.PlayerLeagueGroup{
		Members: make([]utility.LeagueGroupMember, 0, len(lg.Members)),
	}
	for _, m := range lg.Members {
		result.Members = append(result.Members, utility.LeagueGroupMember{
			AttackLoseCount:  m.AttackLoseCount,
			AttackWinCount:   m.AttackWinCount,
			DefenseLoseCount: m.DefenseLoseCount,
			DefenseWinCount:  m.DefenseWinCount,
			LeagueTrophies:   m.LeagueTrophies,
			ClanName:         m.ClanName,
			ClanTag:          m.ClanTag,
			PlayerName:       m.PlayerName,
			PlayerTag:        m.PlayerTag,
		})
	}
	for _, e := range lg.AttackLogs {
		result.AttackLogs = append(result.AttackLogs, utility.LeagueBattleLogEntry{
			CreationTime:          e.CreationTime,
			DestructionPercentage: e.DestructionPercentage,
			OpponentName:          e.OpponentName,
			OpponentPlayerTag:     e.OpponentPlayerTag,
			Stars:                 e.Stars,
			Trophies:              e.Trophies,
		})
	}
	for _, e := range lg.DefenseLogs {
		result.DefenseLogs = append(result.DefenseLogs, utility.LeagueBattleLogEntry{
			CreationTime:          e.CreationTime,
			DestructionPercentage: e.DestructionPercentage,
			OpponentName:          e.OpponentName,
			OpponentPlayerTag:     e.OpponentPlayerTag,
			Stars:                 e.Stars,
			Trophies:              e.Trophies,
		})
	}
	return result
}

func mapCocapiUtilityError(err error, notFoundCode string) error {
	switch {
	case errors.Is(err, cocapi.ErrNotFound):
		if notFoundCode != "" {
			return dmerrors.New(notFoundCode, err.Error())
		}
		return err
	case errors.Is(err, cocapi.ErrInvalidTag):
		return dmerrors.New(dmerrors.ErrCodeInvalidTag, err.Error())
	case errors.Is(err, cocapi.ErrAPINotConfigured):
		return dmerrors.New(dmerrors.ErrCodeAPINotConfigured, err.Error())
	case errors.Is(err, cocapi.ErrAPIAccessDenied):
		return dmerrors.New(dmerrors.ErrCodeAPIAccessDenied, err.Error())
	case errors.Is(err, cocapi.ErrAPIRequestFailed):
		return dmerrors.New(dmerrors.ErrCodeAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrRateLimited):
		return dmerrors.New(dmerrors.ErrCodeAPIRequestFailed, err.Error())
	case errors.Is(err, cocapi.ErrAPIResponseInvalid):
		return dmerrors.New(dmerrors.ErrCodeAPIResponseInvalid, err.Error())
	default:
		return err
	}
}
