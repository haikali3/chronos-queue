package redis

import (
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Lock struct {
	client *goredis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewLock(client *goredis.Client, key string, value string, ttl time.Duration) *Lock {
	return &Lock{
		client: client,
		key:    key,
		value:  value,
		ttl:    ttl,
	}
}
