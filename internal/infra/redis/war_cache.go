package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
)

type WarCache struct {
	client *goredis.Client
}

func NewWarCache(client *goredis.Client) *WarCache {
	return &WarCache{client: client}
}

func (c *WarCache) GetCurrentWar(ctx context.Context, clanTag string) (wardomain.Snapshot, bool, error) {
	if c == nil || c.client == nil {
		return wardomain.Snapshot{}, false, nil
	}
	data, err := c.client.Get(ctx, currentWarKey(clanTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return wardomain.Snapshot{}, false, nil
		}
		return wardomain.Snapshot{}, false, err
	}
	var snapshot wardomain.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return wardomain.Snapshot{}, false, err
	}
	return snapshot, true, nil
}

func (c *WarCache) SetCurrentWar(ctx context.Context, clanTag string, snapshot wardomain.Snapshot, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, currentWarKey(clanTag), data, ttl).Err()
}

func currentWarKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "war:current:" + tag
}


func (c *WarCache) GetCWLGroup(ctx context.Context, clanTag string) (wardomain.CWLGroup, bool, error) {
	if c == nil || c.client == nil {
		return wardomain.CWLGroup{}, false, nil
	}
	data, err := c.client.Get(ctx, cwlGroupKey(clanTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return wardomain.CWLGroup{}, false, nil
		}
		return wardomain.CWLGroup{}, false, err
	}
	var group wardomain.CWLGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return wardomain.CWLGroup{}, false, err
	}
	return group, true, nil
}

func (c *WarCache) SetCWLGroup(ctx context.Context, clanTag string, group wardomain.CWLGroup, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cwlGroupKey(clanTag), data, ttl).Err()
}

func cwlGroupKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "war:cwl:" + tag
}