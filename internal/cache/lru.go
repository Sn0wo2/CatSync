package cache

import lru "github.com/hashicorp/golang-lru/v2"

type LRU[K comparable, V any] struct {
	inner *lru.Cache[K, V]
}

func NewLRU[K comparable, V any](size int) (*LRU[K, V], error) {
	inner, err := lru.New[K, V](size)
	if err != nil {
		return nil, err
	}

	return &LRU[K, V]{inner: inner}, nil
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	return c.inner.Get(key)
}

func (c *LRU[K, V]) Add(key K, value V) {
	c.inner.Add(key, value)
}
