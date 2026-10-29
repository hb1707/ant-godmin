package wxcache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore(0)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := s.Get(fmt.Sprintf("app-%d", i%4))
			c.Set("access_token", i, time.Minute)
			c.Get("access_token")
			c.IsExist("access_token")
			c.Delete("access_token")
		}(i)
	}
	wg.Wait()
}

func TestStoreReusesUntilTTL(t *testing.T) {
	s := NewStore(time.Hour)
	first := s.Get("app")
	if s.Get("app") != first {
		t.Fatal("ttl 内同一 appid 应复用同一份缓存")
	}
	if s.Get("other") == first {
		t.Fatal("不同 appid 不应共用缓存")
	}

	expired := NewStore(time.Nanosecond)
	old := expired.Get("app")
	time.Sleep(time.Millisecond)
	if expired.Get("app") == old {
		t.Fatal("超过 ttl 应换新缓存")
	}
}
