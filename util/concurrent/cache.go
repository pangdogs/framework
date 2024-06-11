package concurrent

import (
	"context"
	"git.golaxy.org/core/utils/types"
	"sync"
	"sync/atomic"
	"time"
)

func NewCache[K comparable, V any]() *Cache[K, V] {
	cache := &Cache[K, V]{
		items: make(map[K]*_CacheItem[K, V]),
	}
	return cache
}

type _CacheItem[K comparable, V any] struct {
	key        K
	value      V
	revision   int64
	ttl        time.Duration
	expireNano atomic.Int64
}

type Cache[K comparable, V any] struct {
	mutex sync.RWMutex
	items map[K]*_CacheItem[K, V]
}

func (c *Cache[K, V]) Set(k K, v V, revision int64, ttl time.Duration) {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.items[k]
	if ok {
		if now < item.expireNano.Load() && revision <= item.revision {
			return
		}
	}

	item = &_CacheItem[K, V]{
		key:      k,
		value:    v,
		revision: revision,
		ttl:      ttl,
	}
	item.expireNano.Store(now + ttl.Nanoseconds())

	c.items[k] = item
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mutex.RLock()
	item, ok := c.items[k]
	c.mutex.RUnlock()
	if !ok {
		return types.ZeroT[V](), false
	}

	now := time.Now().UnixNano()

	expireNano := item.expireNano.Load()
	if now >= expireNano {
		return types.ZeroT[V](), false
	}
	item.expireNano.CompareAndSwap(expireNano, now+item.ttl.Nanoseconds())

	return item.value, true
}

func (c *Cache[K, V]) Del(k K, revision int64) {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.items[k]
	if !ok {
		return
	}

	if now >= item.expireNano.Load() || revision <= item.revision {
		return
	}

	delete(c.items, k)
}

func (c *Cache[K, V]) Clean(num int) {
	if num <= 0 {
		num = 64
	}

	now := time.Now().UnixNano()
	expireItems := make([]*_CacheItem[K, V], 0, num)

	c.mutex.RLock()
	for _, item := range c.items {
		if len(expireItems) >= num {
			break
		}
		if now >= item.expireNano.Load() {
			expireItems = append(expireItems, item)
		}
	}
	c.mutex.RUnlock()

	for _, item := range expireItems {
		c.mutex.Lock()
		existed, _ := c.items[item.key]
		if existed == item {
			delete(c.items, item.key)
		}
		c.mutex.Unlock()
	}
}

func (c *Cache[K, V]) AutoClean(ctx context.Context, interval time.Duration, num int) {
	if ctx == nil {
		ctx = context.Background()
	}

	if interval <= 0 {
		interval = time.Minute
	}

	if num <= 0 {
		num = 64
	}

	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				c.Clean(num)
			}
		}
	}()
}
