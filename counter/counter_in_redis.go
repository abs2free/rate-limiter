package counter

import (
	"context"
	"fmt"
	"log"
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

func ExampleCounterRedis() {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "123456",         // Redis密码，如果没有密码则为空
		DB:       0,                // Redis数据库索引，默认为0
	})

	cr := NewCounterRedis(client, "test", 5)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		allowed, err := cr.Allow(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("第%d次，允许吗?%t \n", i+1, allowed)

		if i%5 == 0 {
			time.Sleep(time.Millisecond * 200)
			fmt.Printf("time.sleep %d \n", time.Millisecond*200)
		}
	}
}
