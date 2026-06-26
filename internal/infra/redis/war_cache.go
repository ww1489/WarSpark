package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	wardomain "github.com/ww1489/WarSpark/internal/domain/war"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

type WarCache struct {
	client        *goredis.Client
	currentWarTTL time.Duration
	cwlGroupTTL   time.Duration
	warLogTTL     time.Duration
	cwlWarTTL     time.Duration
}

func NewWarCache(client *goredis.Client, currentWarTTL, cwlGroupTTL, warLogTTL, cwlWarTTL time.Duration) *WarCache {
	return &WarCache{
		client:        client,
		currentWarTTL: currentWarTTL,
		cwlGroupTTL:   cwlGroupTTL,
		warLogTTL:     warLogTTL,
		cwlWarTTL:     cwlWarTTL,
	}
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

func (c *WarCache) SetCurrentWar(ctx context.Context, clanTag string, snapshot wardomain.Snapshot) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, currentWarKey(clanTag), data, c.currentWarTTL).Err()
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

func (c *WarCache) SetCWLGroup(ctx context.Context, clanTag string, group wardomain.CWLGroup) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cwlGroupKey(clanTag), data, c.cwlGroupTTL).Err()
}

func cwlGroupKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "war:cwl:" + tag
}

func (c *WarCache) GetWarLog(ctx context.Context, clanTag string) (cocapi.ClanWarLogResponse, bool, error) {
	if c == nil || c.client == nil {
		return cocapi.ClanWarLogResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, warLogKey(clanTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return cocapi.ClanWarLogResponse{}, false, nil
		}
		return cocapi.ClanWarLogResponse{}, false, err
	}
	var resp cocapi.ClanWarLogResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return cocapi.ClanWarLogResponse{}, false, err
	}
	return resp, true, nil
}

func (c *WarCache) SetWarLog(ctx context.Context, clanTag string, resp cocapi.ClanWarLogResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, warLogKey(clanTag), data, c.warLogTTL).Err()
}

func warLogKey(clanTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(clanTag))
	tag = strings.TrimPrefix(tag, "#")
	return "war:log:" + tag
}

func (c *WarCache) GetCWLWar(ctx context.Context, warTag string) (cocapi.ClanWar, bool, error) {
	if c == nil || c.client == nil {
		return cocapi.ClanWar{}, false, nil
	}
	data, err := c.client.Get(ctx, cwlWarKey(warTag)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return cocapi.ClanWar{}, false, nil
		}
		return cocapi.ClanWar{}, false, err
	}
	var resp cocapi.ClanWar
	if err := json.Unmarshal(data, &resp); err != nil {
		return cocapi.ClanWar{}, false, err
	}
	return resp, true, nil
}

func (c *WarCache) SetCWLWar(ctx context.Context, warTag string, resp cocapi.ClanWar) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cwlWarKey(warTag), data, c.cwlWarTTL).Err()
}

func cwlWarKey(warTag string) string {
	tag := strings.ToUpper(strings.TrimSpace(warTag))
	tag = strings.TrimPrefix(tag, "#")
	return "war:cwl_war:" + tag
}
