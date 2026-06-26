package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/ranking"
)

type RankingCache struct {
	client *goredis.Client
}

func NewRankingCache(client *goredis.Client) *RankingCache {
	return &RankingCache{client: client}
}

func (c *RankingCache) GetLocations(ctx context.Context) (ranking.LocationListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.LocationListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:locations").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.LocationListResponse{}, false, nil
		}
		return ranking.LocationListResponse{}, false, err
	}
	var resp ranking.LocationListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.LocationListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetLocations(ctx context.Context, resp ranking.LocationListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:locations", data, ttl).Err()
}

func (c *RankingCache) GetClanRanking(ctx context.Context, locationID string) (ranking.ClanRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.ClanRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:clan:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.ClanRankingListResponse{}, false, nil
		}
		return ranking.ClanRankingListResponse{}, false, err
	}
	var resp ranking.ClanRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.ClanRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetClanRanking(ctx context.Context, locationID string, resp ranking.ClanRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:clan:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetPlayerRanking(ctx context.Context, locationID string) (ranking.PlayerRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.PlayerRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:player:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.PlayerRankingListResponse{}, false, nil
		}
		return ranking.PlayerRankingListResponse{}, false, err
	}
	var resp ranking.PlayerRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.PlayerRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetPlayerRanking(ctx context.Context, locationID string, resp ranking.PlayerRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:player:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetClanCapitalRanking(ctx context.Context, locationID string) (ranking.ClanCapitalRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.ClanCapitalRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:capital:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.ClanCapitalRankingListResponse{}, false, nil
		}
		return ranking.ClanCapitalRankingListResponse{}, false, err
	}
	var resp ranking.ClanCapitalRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.ClanCapitalRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetClanCapitalRanking(ctx context.Context, locationID string, resp ranking.ClanCapitalRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:capital:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetClanBuilderBaseRanking(ctx context.Context, locationID string) (ranking.ClanBuilderBaseRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.ClanBuilderBaseRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:builder-clan:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.ClanBuilderBaseRankingListResponse{}, false, nil
		}
		return ranking.ClanBuilderBaseRankingListResponse{}, false, err
	}
	var resp ranking.ClanBuilderBaseRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.ClanBuilderBaseRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetClanBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.ClanBuilderBaseRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:builder-clan:"+locationID, data, ttl).Err()
}

func (c *RankingCache) GetPlayerBuilderBaseRanking(ctx context.Context, locationID string) (ranking.PlayerBuilderBaseRankingListResponse, bool, error) {
	if c == nil || c.client == nil {
		return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "ranking:builder-player:"+locationID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return ranking.PlayerBuilderBaseRankingListResponse{}, false, nil
		}
		return ranking.PlayerBuilderBaseRankingListResponse{}, false, err
	}
	var resp ranking.PlayerBuilderBaseRankingListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return ranking.PlayerBuilderBaseRankingListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *RankingCache) SetPlayerBuilderBaseRanking(ctx context.Context, locationID string, resp ranking.PlayerBuilderBaseRankingListResponse, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "ranking:builder-player:"+locationID, data, ttl).Err()
}
