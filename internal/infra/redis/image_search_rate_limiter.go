package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

const (
	defaultUploadPerMinute = 5
	defaultUploadPerHour   = 50
)

type ImageSearchUploadRateLimiter struct {
	client    *goredis.Client
	perMinute int
	perHour   int
	now       func() time.Time
}

func NewImageSearchUploadRateLimiter(client *goredis.Client) *ImageSearchUploadRateLimiter {
	return &ImageSearchUploadRateLimiter{
		client:    client,
		perMinute: defaultUploadPerMinute,
		perHour:   defaultUploadPerHour,
		now:       time.Now,
	}
}

func (l *ImageSearchUploadRateLimiter) AllowUpload(ctx context.Context, clientID string) error {
	if l == nil || l.client == nil {
		return nil
	}
	if strings.TrimSpace(clientID) == "" {
		clientID = "anonymous"
	}

	now := l.now().UTC()
	minuteKey := uploadRateKey(clientID, "minute", now.Format("200601021504"))
	hourKey := uploadRateKey(clientID, "hour", now.Format("2006010215"))

	minuteCount, err := incrementWithTTL(ctx, l.client, minuteKey, 2*time.Minute)
	if err != nil {
		return err
	}
	hourCount, err := incrementWithTTL(ctx, l.client, hourKey, 2*time.Hour)
	if err != nil {
		return err
	}
	if minuteCount > int64(l.perMinute) || hourCount > int64(l.perHour) {
		return imagesearchdomain.NewUploadError(imagesearchdomain.ErrorRateLimited, "upload rate limited")
	}
	return nil
}

func incrementWithTTL(ctx context.Context, client *goredis.Client, key string, ttl time.Duration) (int64, error) {
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := client.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func uploadRateKey(clientID string, window string, bucket string) string {
	hash := sha256.Sum256([]byte(clientID))
	return fmt.Sprintf("image-search:upload-rate:%s:%s:%s", window, hex.EncodeToString(hash[:]), bucket)
}
