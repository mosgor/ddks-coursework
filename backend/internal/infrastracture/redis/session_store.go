package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TokenBlacklistPrefix = "blacklist:"
	TokenTTL             = 24 * time.Hour
)

func (c *Client) BlacklistToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, token)
	return c.Set(ctx, key, "1", TokenTTL).Err()
}

func (c *Client) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, token)
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}
