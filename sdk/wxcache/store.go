// Package wxcache 为微信 SDK 提供按 appid 复用、可被多个 goroutine 同时使用的内存缓存。
package wxcache

import (
	"sync"
	"time"

	"github.com/silenceper/wechat/v2/cache"
)

// Store 按 appid 复用缓存。日志告警、HTTP 请求会同时取缓存，取用与重建必须在同一把锁内完成。
type Store struct {
	ttl   time.Duration
	mu    sync.Mutex
	items map[string]entry
}

type entry struct {
	cache     cache.Cache
	createdAt time.Time
}

// NewStore ttl 大于 0 时，缓存存活超过 ttl 后整份换新；为 0 时一直复用。
func NewStore(ttl time.Duration) *Store {
	return &Store{ttl: ttl, items: make(map[string]entry)}
}

// Get 返回 appid 对应的缓存，不存在或已过期时新建。
func (s *Store) Get(appid string) cache.Cache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[appid]; ok && (s.ttl <= 0 || time.Since(e.createdAt) <= s.ttl) {
		return e.cache
	}
	e := entry{cache: &lockedCache{inner: cache.NewMemory()}, createdAt: time.Now()}
	s.items[appid] = e
	return e.cache
}

// lockedCache 给每次读写都加锁：wechat 的 cache.Memory 只在 Set/Delete 加锁，
// Get/IsExist 直接读 map，与并发的 Set 相遇会触发 concurrent map read and map write 使进程退出。
type lockedCache struct {
	mu    sync.Mutex
	inner cache.Cache
}

func (c *lockedCache) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Get(key)
}

func (c *lockedCache) Set(key string, val interface{}, timeout time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Set(key, val, timeout)
}

func (c *lockedCache) IsExist(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.IsExist(key)
}

func (c *lockedCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner.Delete(key)
}
