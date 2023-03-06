package limiter

import (
	"context"
	"log"
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
		Addr:         "localhost",
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

func (l *RedisRateLimiter) Allow(key string) bool {
	ctx := context.Background()
	count, err := l.client.Incr(ctx, l.key(key)).Result()
	if err != nil {
		return true // allow request if Redis is down or there's an error
	}
	if count == 1 {
		// set expiration time on the key
		if err := l.client.Expire(ctx, l.key(key), l.duration).Err(); err != nil {
			return true
		}
	}
	log.Println(count, int64(l.limit))
	return count <= int64(l.limit)
}
