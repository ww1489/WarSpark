package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/capital"
)

type CapitalCache struct {
	client        *goredis.Client
	raidSeasonTTL time.Duration
	leagueTTL     time.Duration
}

func NewCapitalCache(client *goredis.Client, raidSeasonTTL, leagueTTL time.Duration) *CapitalCache {
	return &CapitalCache{client: client, raidSeasonTTL: raidSeasonTTL, leagueTTL: leagueTTL}
}

func (c *CapitalCache) GetCapitalRaidSeasons(ctx context.Context, clanTag string) (capital.CapitalRaidSeasonListResponse, bool, error) {
	if c == nil || c.client == nil {
		return capital.CapitalRaidSeasonListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:capital:raid_seasons:"+clanTag).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return capital.CapitalRaidSeasonListResponse{}, false, nil
		}
		return capital.CapitalRaidSeasonListResponse{}, false, err
	}
	var resp capital.CapitalRaidSeasonListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return capital.CapitalRaidSeasonListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *CapitalCache) SetCapitalRaidSeasons(ctx context.Context, clanTag string, resp capital.CapitalRaidSeasonListResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:capital:raid_seasons:"+clanTag, data, c.raidSeasonTTL).Err()
}

func (c *CapitalCache) GetCapitalLeagues(ctx context.Context) (capital.CapitalLeagueListResponse, bool, error) {
	if c == nil || c.client == nil {
		return capital.CapitalLeagueListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:capital:leagues").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return capital.CapitalLeagueListResponse{}, false, nil
		}
		return capital.CapitalLeagueListResponse{}, false, err
	}
	var resp capital.CapitalLeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return capital.CapitalLeagueListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *CapitalCache) SetCapitalLeagues(ctx context.Context, resp capital.CapitalLeagueListResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:capital:leagues", data, c.leagueTTL).Err()
}

func (c *CapitalCache) GetCapitalLeague(ctx context.Context, leagueID string) (capital.CapitalLeague, bool, error) {
	if c == nil || c.client == nil {
		return capital.CapitalLeague{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:capital:league:"+leagueID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return capital.CapitalLeague{}, false, nil
		}
		return capital.CapitalLeague{}, false, err
	}
	var resp capital.CapitalLeague
	if err := json.Unmarshal(data, &resp); err != nil {
		return capital.CapitalLeague{}, false, err
	}
	return resp, true, nil
}

func (c *CapitalCache) SetCapitalLeague(ctx context.Context, leagueID string, resp capital.CapitalLeague) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:capital:league:"+leagueID, data, c.leagueTTL).Err()
}

func (c *CapitalCache) GetBuilderBaseLeagues(ctx context.Context) (capital.BuilderBaseLeagueListResponse, bool, error) {
	if c == nil || c.client == nil {
		return capital.BuilderBaseLeagueListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:builder:leagues").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return capital.BuilderBaseLeagueListResponse{}, false, nil
		}
		return capital.BuilderBaseLeagueListResponse{}, false, err
	}
	var resp capital.BuilderBaseLeagueListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return capital.BuilderBaseLeagueListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *CapitalCache) SetBuilderBaseLeagues(ctx context.Context, resp capital.BuilderBaseLeagueListResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:builder:leagues", data, c.leagueTTL).Err()
}

func (c *CapitalCache) GetBuilderBaseLeague(ctx context.Context, leagueID string) (capital.BuilderBaseLeague, bool, error) {
	if c == nil || c.client == nil {
		return capital.BuilderBaseLeague{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:builder:league:"+leagueID).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return capital.BuilderBaseLeague{}, false, nil
		}
		return capital.BuilderBaseLeague{}, false, err
	}
	var resp capital.BuilderBaseLeague
	if err := json.Unmarshal(data, &resp); err != nil {
		return capital.BuilderBaseLeague{}, false, err
	}
	return resp, true, nil
}

func (c *CapitalCache) SetBuilderBaseLeague(ctx context.Context, leagueID string, resp capital.BuilderBaseLeague) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:builder:league:"+leagueID, data, c.leagueTTL).Err()
}
