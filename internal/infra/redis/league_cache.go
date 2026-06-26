package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/league"
)

type LeagueCache struct {
	client *goredis.Client
}

func NewLeagueCache(client *goredis.Client) *LeagueCache {
	return &LeagueCache{client: client}
}

func (c *LeagueCache) GetLeagues(ctx context.Context) (league.LeagueListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:list").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueListResponse{}, false, nil
		}
		return league.LeagueListResponse{}, false, err
	}
	var resp league.LeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.LeagueListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetLeagues(ctx context.Context, resp league.LeagueListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:list", data, ttl).Err()
}

func (c *LeagueCache) GetLeague(ctx context.Context, id string) (league.League, bool, error) {
	if c == nil || c.client == nil {
		return league.League{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:"+id).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.League{}, false, nil
		}
		return league.League{}, false, err
	}
	var l league.League
	if err := json.Unmarshal(data, &l); err != nil {
		return league.League{}, false, err
	}
	return l, true, nil
}

func (c *LeagueCache) SetLeague(ctx context.Context, id string, l league.League, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:"+id, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueSeasons(ctx context.Context, leagueID string) (league.LeagueSeasonListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueSeasonListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:seasons:"+leagueID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueSeasonListResponse{}, false, nil
		}
		return league.LeagueSeasonListResponse{}, false, err
	}
	var resp league.LeagueSeasonListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.LeagueSeasonListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueSeasons(ctx context.Context, leagueID string, resp league.LeagueSeasonListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:seasons:"+leagueID, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueSeasonRankings(ctx context.Context, leagueID, season string) (league.LeagueSeasonRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueSeasonRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:rankings:"+leagueID+":"+season).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueSeasonRankingListResponse{}, false, nil
		}
		return league.LeagueSeasonRankingListResponse{}, false, err
	}
	var resp league.LeagueSeasonRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.LeagueSeasonRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueSeasonRankings(ctx context.Context, leagueID, season string, resp league.LeagueSeasonRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:rankings:"+leagueID+":"+season, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueTiers(ctx context.Context, leagueID, season string) (league.LeagueTierListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueTierListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:tiers:"+leagueID+":"+season).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueTierListResponse{}, false, nil
		}
		return league.LeagueTierListResponse{}, false, err
	}
	var resp league.LeagueTierListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.LeagueTierListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueTiers(ctx context.Context, leagueID, season string, resp league.LeagueTierListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:tiers:"+leagueID+":"+season, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueTier(ctx context.Context, tierID string) (league.LeagueTier, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueTier{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:tier:"+tierID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueTier{}, false, nil
		}
		return league.LeagueTier{}, false, err
	}
	var t league.LeagueTier
	if err := json.Unmarshal(data, &t); err != nil {
		return league.LeagueTier{}, false, err
	}
	return t, true, nil
}

func (c *LeagueCache) SetLeagueTier(ctx context.Context, tierID string, t league.LeagueTier, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:tier:"+tierID, data, ttl).Err()
}

func (c *LeagueCache) GetLeagueHistory(ctx context.Context, playerTag string) (league.LeagueSeasonResultListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.LeagueSeasonResultListResponse{}, false, nil
	}
	tag := normalizePlayerTagForCache(playerTag)
	data, err := c.client.Get(ctx, "league:history:"+tag).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.LeagueSeasonResultListResponse{}, false, nil
		}
		return league.LeagueSeasonResultListResponse{}, false, err
	}
	var resp league.LeagueSeasonResultListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.LeagueSeasonResultListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetLeagueHistory(ctx context.Context, playerTag string, resp league.LeagueSeasonResultListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	tag := normalizePlayerTagForCache(playerTag)
	return c.client.Set(ctx, "league:history:"+tag, data, ttl).Err()
}

func (c *LeagueCache) GetWarLeagues(ctx context.Context) (league.WarLeagueListResponse, bool, error) {
	if c == nil || c.client == nil {
		return league.WarLeagueListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:war-list").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.WarLeagueListResponse{}, false, nil
		}
		return league.WarLeagueListResponse{}, false, err
	}
	var resp league.WarLeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return league.WarLeagueListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LeagueCache) SetWarLeagues(ctx context.Context, resp league.WarLeagueListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:war-list", data, ttl).Err()
}

func (c *LeagueCache) GetWarLeague(ctx context.Context, id string) (league.WarLeague, bool, error) {
	if c == nil || c.client == nil {
		return league.WarLeague{}, false, nil
	}
	data, err := c.client.Get(ctx, "league:war:"+id).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return league.WarLeague{}, false, nil
		}
		return league.WarLeague{}, false, err
	}
	var l league.WarLeague
	if err := json.Unmarshal(data, &l); err != nil {
		return league.WarLeague{}, false, err
	}
	return l, true, nil
}

func (c *LeagueCache) SetWarLeague(ctx context.Context, id string, l league.WarLeague, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "league:war:"+id, data, ttl).Err()
}

func normalizePlayerTagForCache(tag string) string {
	tag = strings.ToUpper(strings.TrimSpace(tag))
	tag = strings.TrimPrefix(tag, "#")
	return tag
}
