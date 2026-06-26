package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/domain/label"
)

type LabelCache struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewLabelCache(client *goredis.Client, ttl time.Duration) *LabelCache {
	return &LabelCache{client: client, ttl: ttl}
}

func (c *LabelCache) GetClanLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	if c == nil || c.client == nil {
		return label.LabelListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "label:clan").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return label.LabelListResponse{}, false, nil
		}
		return label.LabelListResponse{}, false, err
	}
	var resp label.LabelListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return label.LabelListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LabelCache) SetClanLabels(ctx context.Context, resp label.LabelListResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "label:clan", data, c.ttl).Err()
}

func (c *LabelCache) GetPlayerLabels(ctx context.Context) (label.LabelListResponse, bool, error) {
	if c == nil || c.client == nil {
		return label.LabelListResponse{}, false, nil
	}
	data, err := c.client.Get(ctx, "label:player").Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return label.LabelListResponse{}, false, nil
		}
		return label.LabelListResponse{}, false, err
	}
	var resp label.LabelListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return label.LabelListResponse{}, false, err
	}
	return resp, true, nil
}

func (c *LabelCache) SetPlayerLabels(ctx context.Context, resp label.LabelListResponse) error {
	if c == nil || c.client == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "label:player", data, c.ttl).Err()
}
