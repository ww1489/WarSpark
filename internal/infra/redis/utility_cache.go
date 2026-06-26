package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/utility"
)

type UtilityCache struct {
	client      *goredis.Client
	goldPassTTL time.Duration
}

func NewUtilityCache(client *goredis.Client, goldPassTTL time.Duration) *UtilityCache {
	return &UtilityCache{client: client, goldPassTTL: goldPassTTL}
}

func (c *UtilityCache) GetGoldPass(ctx context.Context) (utility.GoldPassSeason, bool, error) {
	if c == nil || c.client == nil {
		return utility.GoldPassSeason{}, false, nil
	}
	data, err := c.client.Get(ctx, "wsp:goldpass:current").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return utility.GoldPassSeason{}, false, nil
		}
		return utility.GoldPassSeason{}, false, err
	}
	var resp utility.GoldPassSeason
	if err := json.Unmarshal(data, &resp); err != nil {
		return utility.GoldPassSeason{}, false, err
	}
	return resp, true, nil
}

func (c *UtilityCache) SetGoldPass(ctx context.Context, resp utility.GoldPassSeason) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "wsp:goldpass:current", data, c.goldPassTTL).Err()
}
