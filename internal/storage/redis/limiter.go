package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client    *goredis.Client
	keyPrefix string
	limit     int
	window    time.Duration
}

func NewRateLimiter(client *goredis.Client, keyPrefix string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client:    client,
		keyPrefix: keyPrefix,
		limit:     limit,
		window:    window,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := r.keyPrefix + ":" + key
	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.client.Expire(ctx, redisKey, r.window)
	}
	return count <= int64(r.limit), nil

}
