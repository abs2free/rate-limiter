package ratelimiter_test

import (
	"log"
	ratelimiter "rate-limiter"
	"sync"
	"time"
)

func ExampleCounter() {
	var lr ratelimiter.Counter
	lr.Set(3, time.Second) // 1s内最多请求3次

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			if lr.Allow() {
				log.Println("ok:", i)
			} else {
				log.Println("fail:", i)
			}
			wg.Done()
		}(i)
		time.Sleep(200 * time.Millisecond)
	}
	wg.Wait()
}
