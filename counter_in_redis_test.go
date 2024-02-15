package ratelimiter_test

import (
	"context"
	"fmt"
	"log"
	ratelimiter "rate-limiter"
	"time"

	"github.com/redis/go-redis/v9"
)

func ExampleCounterRedis() {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "123456",         // Redis密码，如果没有密码则为空
		DB:       0,                // Redis数据库索引，默认为0
	})

	cr := ratelimiter.NewCounterRedis(client, "test", 5)
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
