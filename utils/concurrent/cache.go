/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package concurrent

import (
	"context"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CacheDefaultCleanNum      = 64
	CacheDefaultCleanInterval = 30 * time.Second
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
	mutex        sync.RWMutex
	items        map[K]*_CacheItem[K, V]
	onAdd, onDel generic.Action2[K, V]
}

func (c *Cache[K, V]) Set(k K, v V, revision int64, ttl time.Duration) V {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.items[k]
	if ok {
		if revision <= item.revision {
			return item.value
		}
		if item.ttl > 0 && now < item.expireNano.Load() {
			return item.value
		}
		c.onDel.UnsafeCall(k, item.value)
	}

	item = &_CacheItem[K, V]{
		key:      k,
		value:    v,
		revision: revision,
		ttl:      ttl,
	}
	if ttl > 0 {
		item.expireNano.Store(now + ttl.Nanoseconds())
	}

	c.items[k] = item
	c.onAdd.UnsafeCall(k, v)
	return v
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mutex.RLock()
	item, ok := c.items[k]
	c.mutex.RUnlock()
	if !ok {
		return types.ZeroT[V](), false
	}

	if item.ttl > 0 {
		now := time.Now().UnixNano()

		expireNano := item.expireNano.Load()
		if now >= expireNano {
			return types.ZeroT[V](), false
		}
		item.expireNano.CompareAndSwap(expireNano, now+item.ttl.Nanoseconds())
	}

	return item.value, true
}

func (c *Cache[K, V]) Snapshot() generic.UnorderedSliceMap[K, V] {
	now := time.Now().UnixNano()

	c.mutex.RLock()
	defer c.mutex.RLock()

	sliceMap := make(generic.UnorderedSliceMap[K, V], 0, len(c.items))

	for _, item := range c.items {
		if item.ttl > 0 && now >= item.expireNano.Load() {
			continue
		}
		sliceMap.Add(item.key, item.value)
	}

	return sliceMap
}

func (c *Cache[K, V]) Del(k K, revision int64) {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, ok := c.items[k]
	if !ok {
		return
	}

	if revision <= item.revision {
		return
	}

	if item.ttl > 0 && now >= item.expireNano.Load() {
		return
	}

	delete(c.items, k)
	c.onDel.UnsafeCall(k, item.value)
}

func (c *Cache[K, V]) RefreshTTL(k K) {
	c.mutex.RLock()
	item, ok := c.items[k]
	c.mutex.RUnlock()
	if ok {
		if item.ttl > 0 {
			now := time.Now().UnixNano()

			expireNano := item.expireNano.Load()
			if now >= expireNano {
				return
			}
			item.expireNano.CompareAndSwap(expireNano, now+item.ttl.Nanoseconds())
		}
	}
}

func (c *Cache[K, V]) Clean(num int) {
	if num <= 0 {
		num = CacheDefaultCleanNum
	}

	now := time.Now().UnixNano()
	expireItems := make([]*_CacheItem[K, V], 0, num)

	c.mutex.RLock()
	for _, item := range c.items {
		if len(expireItems) >= num {
			break
		}
		if item.ttl > 0 && now >= item.expireNano.Load() {
			expireItems = append(expireItems, item)
		}
	}
	c.mutex.RUnlock()

	for _, item := range expireItems {
		func() {
			c.mutex.Lock()
			defer c.mutex.Unlock()

			existed, _ := c.items[item.key]
			if existed == item {
				delete(c.items, item.key)
				c.onDel.UnsafeCall(item.key, item.value)
			}
		}()
	}
}

func (c *Cache[K, V]) AutoClean(ctx context.Context, interval time.Duration, num int) {
	if ctx == nil {
		ctx = context.Background()
	}

	if interval <= 0 {
		interval = CacheDefaultCleanInterval
	}

	if num <= 0 {
		num = CacheDefaultCleanNum
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

func (c *Cache[K, V]) OnAdd(cb generic.Action2[K, V]) {
	c.onAdd = cb
}

func (c *Cache[K, V]) OnDel(cb generic.Action2[K, V]) {
	c.onDel = cb
}
