package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func hmget(pipe redis.Pipeliner, key string, fields []string) *redis.SliceCmd {
	slicecmd := pipe.HMGet(ctx, key, fields...)
	return slicecmd
}

type RedisClient struct {
	client *redis.Client
}

var ctx = context.Background()

// TODO: Use config instead
func NewRedisClient(host string) *RedisClient {
	instance := &RedisClient{}
	instance.client = redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     1000,
		PoolTimeout:  2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	})
	return instance
}

func (r *RedisClient) Do(args ...interface{}) *redis.Cmd {
	return r.client.Do(ctx, args...)
}

func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

func (r *RedisClient) ResetCount(key string, field string) (string, error) {
	countKey := fmt.Sprintf(`%s_%s`, key, "counter")
	val, err := r.client.Do(ctx, "HSET", countKey, field, 0).Result()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(val.(int64))), err
}

func (r *RedisClient) NextCount(key string, field string) (string, error) {
	countKey := fmt.Sprintf(`%s_%s`, key, "counter")
	val, err := r.client.Do(ctx, "HINCRBY", countKey, field, 1).Result()
	if err != nil {
		log.Println(err)
		return "", err
	}
	return strconv.Itoa(int(val.(int64))), err
}

func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RedisClient) HGet(key string, field string) (string, error) {
	cmd := r.client.HGet(ctx, key, field)
	result, err := cmd.Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (r *RedisClient) MHMGet(keys map[string][]string, keyPrefix string) (map[string][]interface{}, error) {
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.SliceCmd)
	result := make(map[string][]interface{})
	for key, fields := range keys {
		cmds[key] = pipe.HMGet(ctx, key+keyPrefix, fields...)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range cmds {
		res, _ := v.Result()
		result[k] = res
	}
	return result, nil
}

func (r *RedisClient) MHGetAll(keys []string) (map[string]map[string]string, error) {
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.MapStringStringCmd)
	result := make(map[string]map[string]string)
	for _, key := range keys {
		cmds[key] = pipe.HGetAll(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range cmds {
		res, _ := v.Result()
		result[k] = res
	}
	return result, nil
}

func (r *RedisClient) HSet(key string, field string, value interface{}) error {
	_, err := r.client.Do(ctx, "HSet", key, field, value).Result()
	if err != nil {
		return err
	}
	return nil
}

// TODO: Return correct result
func (r *RedisClient) HMSet(key string, properties map[string]interface{}) error {
	if properties == nil {
		return nil
	}
	values := make([]interface{}, 0)
	for k, v := range properties {
		values = append(values, k, v)
	}
	r.client.HMSet(ctx, key, values...)
	return nil
}

func (r *RedisClient) MHMSet(dataMap map[string]map[string]interface{}, countersMap map[string]map[string]int64, keyPrefix string) error {
	if dataMap == nil {
		return nil
	}
	pipe := r.client.Pipeline()
	for key, props := range dataMap {
		//TODO: Move it to config
		delete(props, "delegate")
		values := make([]interface{}, 0)
		for k, v := range props {
			values = append(values, k, v)
		}
		if len(values) > 0 {
			pipe.HMSet(ctx, key+keyPrefix, values...)
		}
	}
	if countersMap != nil {
		for key, counters := range countersMap {
			for k, v := range counters {
				pipe.HIncrBy(ctx, key+keyPrefix, k, v)
			}
		}
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) Set(key string, value string, expireTime int64) error {
	at := time.Unix(expireTime, 0) //converting Unix to UTC(to Time object)
	now := time.Now()
	err := r.client.Set(ctx, key, value, at.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) Get(key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}
