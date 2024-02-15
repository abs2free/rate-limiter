package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CounterRedisKey = "count-limit-"
)

type CounterRedis struct {
	Redis   *redis.Client
	Key     string
	Counter int64
}

func NewCounterRedis(redis *redis.Client, key string, counter int64) *CounterRedis {
	key = CounterRedisKey + key + time.Now().Format("2006-01-02 15:04:05")

	return &CounterRedis{
		Redis:   redis,
		Key:     key,
		Counter: counter,
	}
}

func (cr *CounterRedis) Allow(ctx context.Context) (bool, error) {
	counter, err := cr.Redis.Incr(ctx, cr.Key).Result()
	if err != nil {
		return false, err
	}

	// 设置过期时间
	if counter == 1 {
		_, err = cr.Redis.Expire(ctx, cr.Key, time.Second).Result()
		if err != nil {
			return false, err
		}
	}

	// 超出计数限制
	if counter > cr.Counter {
		return false, nil
	}

	return true, nil
}
