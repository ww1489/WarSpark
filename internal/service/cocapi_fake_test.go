package service

import (
	"context"

	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type fakeCocapiClient struct {
	locations          cocapi.LocationListResponse
	clanRanking        cocapi.ClanRankingListResponse
	playerRanking      cocapi.PlayerRankingListResponse
	capitalRanking     cocapi.ClanCapitalRankingListResponse
	builderClan        cocapi.ClanBuilderBaseRankingListResponse
	builderPlayer      cocapi.PlayerBuilderBaseRankingListResponse
	leagues            cocapi.LeagueListResponse
	league             cocapi.League
	leagueSeasons      cocapi.LeagueSeasonListResponse
	leagueRankings     cocapi.PlayerRankingListResponse
	leagueTiers        cocapi.LeagueTierListResponse
	leagueTier         cocapi.LeagueTier
	leagueHistory      cocapi.LeagueSeasonResultListResponse
	warLeagues         cocapi.WarLeagueListResponse
	warLeague          cocapi.WarLeague
	clanLabels         cocapi.LabelListResponse
	playerLabels       cocapi.LabelListResponse
	builderBaseLeagues cocapi.BuilderBaseLeagueListResponse
	builderBaseLeague  cocapi.BuilderBaseLeague
	capitalLeagues     cocapi.CapitalLeagueListResponse
	capitalLeague      cocapi.CapitalLeague
	capitalRaidSeasons cocapi.ClanCapitalRaidSeasonsResponse
	err                error
}

func (f *fakeCocapiClient) GetLocations(ctx context.Context, query cocapi.QueryGetLocations) (cocapi.LocationListResponse, error) {
	return f.locations, f.err
}
func (f *fakeCocapiClient) GetClanRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanRanking) (cocapi.ClanRankingListResponse, error) {
	return f.clanRanking, f.err
}
func (f *fakeCocapiClient) GetPlayerRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerRanking) (cocapi.PlayerRankingListResponse, error) {
	return f.playerRanking, f.err
}
func (f *fakeCocapiClient) GetClanCapitalRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanCapitalRanking) (cocapi.ClanCapitalRankingListResponse, error) {
	return f.capitalRanking, f.err
}
func (f *fakeCocapiClient) GetClanBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetClanBuilderBaseRanking) (cocapi.ClanBuilderBaseRankingListResponse, error) {
	return f.builderClan, f.err
}
func (f *fakeCocapiClient) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string, query cocapi.QueryGetPlayerBuilderBaseRanking) (cocapi.PlayerBuilderBaseRankingListResponse, error) {
	return f.builderPlayer, f.err
}
func (f *fakeCocapiClient) GetLeagues(ctx context.Context, query cocapi.QueryGetLeagues) (cocapi.LeagueListResponse, error) {
	return f.leagues, f.err
}
func (f *fakeCocapiClient) GetLeague(ctx context.Context, leagueID string) (cocapi.League, error) {
	return f.league, f.err
}
func (f *fakeCocapiClient) GetLeagueSeasons(ctx context.Context, leagueID string, query cocapi.QueryGetLeagueSeasons) (cocapi.LeagueSeasonListResponse, error) {
	return f.leagueSeasons, f.err
}
func (f *fakeCocapiClient) GetLeagueSeasonRankings(ctx context.Context, leagueID, seasonID string, query cocapi.QueryGetLeagueSeasonRankings) (cocapi.PlayerRankingListResponse, error) {
	return f.leagueRankings, f.err
}
func (f *fakeCocapiClient) GetLeagueTiers(ctx context.Context, query cocapi.QueryGetLeagueTiers) (cocapi.LeagueTierListResponse, error) {
	return f.leagueTiers, f.err
}
func (f *fakeCocapiClient) GetLeagueTier(ctx context.Context, tierID string) (cocapi.LeagueTier, error) {
	return f.leagueTier, f.err
}
func (f *fakeCocapiClient) GetLeagueHistory(ctx context.Context, playerTag string) (cocapi.LeagueSeasonResultListResponse, error) {
	return f.leagueHistory, f.err
}
func (f *fakeCocapiClient) GetWarLeagues(ctx context.Context, query cocapi.QueryGetWarLeagues) (cocapi.WarLeagueListResponse, error) {
	return f.warLeagues, f.err
}
func (f *fakeCocapiClient) GetWarLeague(ctx context.Context, leagueID string) (cocapi.WarLeague, error) {
	return f.warLeague, f.err
}
func (f *fakeCocapiClient) GetClanLabels(ctx context.Context, query cocapi.QueryGetClanLabels) (cocapi.LabelListResponse, error) {
	return f.clanLabels, f.err
}
func (f *fakeCocapiClient) GetPlayerLabels(ctx context.Context, query cocapi.QueryGetPlayerLabels) (cocapi.LabelListResponse, error) {
	return f.playerLabels, f.err
}

func (f *fakeCocapiClient) GetBuilderBaseLeagues(ctx context.Context, query cocapi.QueryGetBuilderBaseLeagues) (cocapi.BuilderBaseLeagueListResponse, error) {
	return f.builderBaseLeagues, f.err
}

func (f *fakeCocapiClient) GetBuilderBaseLeague(ctx context.Context, leagueId string) (cocapi.BuilderBaseLeague, error) {
	return f.builderBaseLeague, f.err
}

func (f *fakeCocapiClient) GetCapitalLeagues(ctx context.Context, query cocapi.QueryGetCapitalLeagues) (cocapi.CapitalLeagueListResponse, error) {
	return f.capitalLeagues, f.err
}

func (f *fakeCocapiClient) GetCapitalLeague(ctx context.Context, leagueId string) (cocapi.CapitalLeague, error) {
	return f.capitalLeague, f.err
}

func (f *fakeCocapiClient) SearchClans(ctx context.Context, query cocapi.QuerySearchClans) (cocapi.ClanListResponse, error) {
	return cocapi.ClanListResponse{}, f.err
}

func (f *fakeCocapiClient) GetCapitalRaidSeasons(ctx context.Context, clanTag string, query cocapi.QueryGetCapitalRaidSeasons) (cocapi.ClanCapitalRaidSeasonsResponse, error) {
	return f.capitalRaidSeasons, f.err
}

func (f *fakeCocapiClient) GetClanWarLog(ctx context.Context, clanTag string, query cocapi.QueryGetClanWarLog) (cocapi.ClanWarLogResponse, error) {
	return cocapi.ClanWarLogResponse{}, nil
}

func (f *fakeCocapiClient) GetClanWarLeagueWar(ctx context.Context, warTag string) (cocapi.ClanWar, error) {
	return cocapi.ClanWar{}, nil
}

func (f *fakeCocapiClient) GetCurrentGoldPassSeason(ctx context.Context) (cocapi.GoldPassSeason, error) {
	return cocapi.GoldPassSeason{}, f.err
}

func (f *fakeCocapiClient) GetLocation(ctx context.Context, locationId string) (cocapi.Location, error) {
	return cocapi.Location{}, f.err
}

func (f *fakeCocapiClient) VerifyToken(ctx context.Context, playerTag string, body cocapi.VerifyTokenRequest) (cocapi.VerifyTokenResponse, error) {
	return cocapi.VerifyTokenResponse{}, f.err
}

func (f *fakeCocapiClient) GetLeagueGroup(ctx context.Context, leagueGroupTag string, leagueSeasonId string, query cocapi.QueryGetLeagueGroup) (cocapi.LeagueGroup, error) {
	return cocapi.LeagueGroup{}, f.err
}
