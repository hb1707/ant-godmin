package qywx

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 复现 Agent 崩溃场景：多条 Error 日志同时推企业微信，各自首次取同一 appid 的缓存并读写 access token。
func TestMemoryConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := Memory(fmt.Sprintf("app-%d", i%4))
			c.Set("access_token", i, time.Minute)
			c.Get("access_token")
			c.IsExist("access_token")
		}(i)
	}
	wg.Wait()
}
