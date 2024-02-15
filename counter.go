package ratelimiter

import (
	"sync"
	"time"
)

type Counter struct {
	rate   int           // 计数周期内最多允许的请求数
	begin  time.Time     // 计数开始时间
	window time.Duration // 计数周期
	count  int           // 计数周期内累计收到的请求数
	lock   sync.Mutex
}

func (l *Counter) Allow() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.count < l.rate-1 {
		//没有达到速率限制，计数加1
		l.count++
		return true
	}

	now := time.Now()
	if now.Sub(l.begin) >= l.window {
		//速度允许范围内， 重置计数器
		l.Reset(now)
		return true
	} else {
		return false
	}
}

func (l *Counter) Set(r int, window time.Duration) {
	l.rate = r
	l.begin = time.Now()
	l.window = window
	l.count = 0
}

func (l *Counter) Reset(t time.Time) {
	l.begin = t
	l.count = 0
}
