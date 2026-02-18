package redis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type Lock struct {
	client *goredis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewLock(client *goredis.Client, key string, ttl time.Duration) *Lock {
	return &Lock{
		client: client,
		key:    key,
		value:  uuid.NewString(),
		ttl:    ttl,
	}
}

func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	result := l.client.SetArgs(ctx, l.key, l.value, goredis.SetArgs{
		Mode: "NX",
		TTL:  l.ttl,
	})

	err := result.Err()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Lock) Release(ctx context.Context) error {
	script := `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
      return redis.call("DEL", KEYS[1])
    else
      return 0
    end
  `
	result := l.client.Eval(ctx, script, []string{l.key}, l.value)
	return result.Err()
}
