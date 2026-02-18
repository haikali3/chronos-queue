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

