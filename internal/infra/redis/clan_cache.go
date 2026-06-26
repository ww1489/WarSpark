package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	clandomain "github.com/ww1489/WarSpark/internal/domain/clan"
)

type ClanCache struct {
	client *goredis.Client
}

func NewClanCache(client *goredis.Client) *ClanCache {
	return &ClanCache{client: client}
}

func (c *ClanCache) GetClan(ctx context.Context, clanTag string) (clandomain.ClanDetail, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.ClanDetail{}, false, nil
	}
	data, err := c.client.Get(ctx, clanDetailKey(clanTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.ClanDetail{}, false, nil
		}
		return clandomain.ClanDetail{}, false, err
	}
	var detail clandomain.ClanDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return clandomain.ClanDetail{}, false, err
	}
	return detail, true, nil
}

func (c *ClanCache) SetClan(ctx context.Context, clanTag string, detail clandomain.ClanDetail, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, clanDetailKey(clanTag), data, ttl).Err()
}

func (c *ClanCache) GetPlayer(ctx context.Context, playerTag string) (clandomain.PlayerOverview, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.PlayerOverview{}, false, nil
	}
	data, err := c.client.Get(ctx, playerOverviewKey(playerTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.PlayerOverview{}, false, nil
		}
		return clandomain.PlayerOverview{}, false, err
	}
	var player clandomain.PlayerOverview
	if err := json.Unmarshal(data, &player); err != nil {
		return clandomain.PlayerOverview{}, false, err
	}
	return player, true, nil
}

func (c *ClanCache) SetPlayer(ctx context.Context, playerTag string, player clandomain.PlayerOverview, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, playerOverviewKey(playerTag), data, ttl).Err()
}

func (c *ClanCache) GetBattleLog(ctx context.Context, playerTag string) (clandomain.BattleLogSummary, bool, error) {
	if c == nil || c.client == nil {
		return clandomain.BattleLogSummary{}, false, nil
	}
	data, err := c.client.Get(ctx, battleLogKey(playerTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return clandomain.BattleLogSummary{}, false, nil
		}
		return clandomain.BattleLogSummary{}, false, err
	}
	var log clandomain.BattleLogSummary
	if err := json.Unmarshal(data, &log); err != nil {
		return clandomain.BattleLogSummary{}, false, err
	}
	return log, true, nil
}

func (c *ClanCache) SetBattleLog(ctx context.Context, playerTag string, log clandomain.BattleLogSummary, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, battleLogKey(playerTag), data, ttl).Err()
}

func clanDetailKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "clan:detail:" + tag
}

func playerOverviewKey(playerTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(playerTag))
	tag = strings.TrimPrefix(tag, "#")
	return "player:overview:" + tag
}

func battleLogKey(playerTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(playerTag))
	tag = strings.TrimPrefix(tag, "#")
	return "player:battlelog:" + tag
}
