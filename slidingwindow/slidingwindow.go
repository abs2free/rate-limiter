package slidingwindow

import (
	"fmt"
	"sync"
	"time"
)

var winMux map[string]*sync.RWMutex

func init() {
	winMux = make(map[string]*sync.RWMutex)
}

type timeSlot struct {
	timestamp time.Time // 这个timeSlot的时间起点
	count     int       // 落在这个timeSlot内的请求数
}

type winInfo []*timeSlot

func (w winInfo) count() int {
	var count int
	for _, ts := range w {
		count += ts.count
	}
	return count
}

func (w winInfo) incr(now time.Time, slotDuration time.Duration) winInfo {
	var lastSlot *timeSlot
	if len(w) > 0 {
		lastSlot = w[len(w)-1]
		if lastSlot.timestamp.Add(slotDuration).Before(now) {
			lastSlot = &timeSlot{timestamp: now, count: 1}
			w = append(w, lastSlot)
		} else {
			lastSlot.count++
		}
	} else {
		lastSlot = &timeSlot{timestamp: now, count: 1}
		w = append(w, lastSlot)
	}

	return w
}

type slidingWindow struct {
	SlotDuration time.Duration // time slot 的长度
	WinDuration  time.Duration // sliding window的长度
	slotNums     int           // window内有多少个slot
	maxReq       int           // window duration内最大的请求数
	windows      map[string]winInfo
}

func NewSlidingWindow(slotDuratin, winDuration time.Duration, maxReq int) *slidingWindow {
	return &slidingWindow{
		SlotDuration: slotDuratin,
		WinDuration:  winDuration,
		slotNums:     int(winDuration / slotDuratin),
		maxReq:       maxReq,
		windows:      make(map[string]winInfo),
	}
}

// getWindow 获取userid/ip的时间窗
func (s *slidingWindow) getWindow(uidOrIp string) winInfo {
	win, ok := s.windows[uidOrIp]
	if !ok {
		win = make(winInfo, 0, s.slotNums)
	}
	return win
}

func (s *slidingWindow) storeWindow(uidOrIp string, win winInfo) {
	s.windows[uidOrIp] = win
}

func (s *slidingWindow) validate(uidOrIp string) bool {
	// 同一个userid/ip并发安全
	mu, ok := winMux[uidOrIp]
	if !ok {
		var m sync.RWMutex
		mu = &m
		winMux[uidOrIp] = mu
	}

	mu.Lock()
	defer mu.Unlock()

	winInfo := s.getWindow(uidOrIp)
	now := time.Now()

	// 已经过期的time slot移出时间窗
	timeoutOffset := -1
	for i, ts := range winInfo {
		if ts.timestamp.Add(s.WinDuration).After(now) {
			break
		}
		timeoutOffset = i
	}
	if timeoutOffset > -1 {
		winInfo = winInfo[timeoutOffset+1:]
	}

	// 判断请求是否超限额
	var result bool
	if winInfo.count() < s.maxReq {
		result = true
	}

	// 记录这次的请求
	// var lastSlot *timeSlot
	// if len(winInfo) > 0 {
	// 	lastSlot = winInfo[len(winInfo)-1]
	// 	if lastSlot.timestamp.Add(s.SlotDuration).Before(now) {
	// 		lastSlot = &timeSlot{timestamp: now, count: 1}
	// 		winInfo = append(winInfo, lastSlot)
	// 	} else {
	// 		lastSlot.count++
	// 	}
	// } else {
	// 	lastSlot = &timeSlot{timestamp: now, count: 1}
	// 	winInfo = append(winInfo, lastSlot)
	// }
	winInfo = winInfo.incr(now, s.SlotDuration)

	s.storeWindow(uidOrIp, winInfo)

	return result
}

func (s *slidingWindow) getUidOrIp() string {
	return "127.0.0.1"
}

func (s *slidingWindow) IsLimited() bool {
	return !s.validate(s.getUidOrIp())
}

func ExampleSlidingWindow() {
	limiter := NewSlidingWindow(100*time.Millisecond, time.Second, 10)
	for i := 0; i < 5; i++ {
		fmt.Println(limiter.IsLimited())
	}
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 5; i++ {
		fmt.Println(limiter.IsLimited())
	}
	fmt.Println(limiter.IsLimited())
	for _, v := range limiter.windows[limiter.getUidOrIp()] {
		fmt.Println(v.timestamp, v.count)
	}

	fmt.Println("a thousand years later...")
	time.Sleep(time.Second)
	for i := 0; i < 7; i++ {
		fmt.Println(limiter.IsLimited())
	}
	for _, v := range limiter.windows[limiter.getUidOrIp()] {
		fmt.Println(v.timestamp, v.count)
	}
}
