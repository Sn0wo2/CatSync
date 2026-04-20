package cache

import "time"

import "github.com/jellydator/ttlcache/v3"

type TTL[K comparable, V any] struct {
	inner *ttlcache.Cache[K, V]
}

func NewTTL[K comparable, V any](ttl time.Duration) *TTL[K, V] {
	c := ttlcache.New(
		ttlcache.WithTTL[K, V](ttl),
	)

	go c.Start()

	return &TTL[K, V]{inner: c}
}

func (c *TTL[K, V]) Get(key K) (V, bool) {
	item := c.inner.Get(key)
	if item != nil {
		return item.Value(), true
	}

	var zero V

	return zero, false
}

func (c *TTL[K, V]) Set(key K, value V) {
	c.inner.Set(key, value, ttlcache.DefaultTTL)
}

func (c *TTL[K, V]) Stop() {
	c.inner.Stop()
}
