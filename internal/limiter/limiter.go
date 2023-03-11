package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client   *redis.Client
	prefix   string
	limit    int
	duration time.Duration
}

func NewRedisRateLimiter(prefix string, limit int, duration time.Duration) *RedisRateLimiter {
	r := &RedisRateLimiter{
		prefix:   prefix,
		limit:    limit,
		duration: duration,
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     1000,
		PoolTimeout:  2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	})
	return r
}

func (l *RedisRateLimiter) key(key string) string {
	return l.prefix + key
}

func (l *RedisRateLimiter) Allow(key uint64) bool {
	ctx := context.Background()
	strKey := fmt.Sprintf(`%d`, key)
	count, err := l.client.Incr(ctx, l.key(strKey)).Result()
	if err != nil {
		return false
	}
	if count == 1 {
		// set expiration time on the key
		if err := l.client.Expire(ctx, l.key(strKey), l.duration).Err(); err != nil {
			return true
		}
	}
	return count <= int64(l.limit)
}
