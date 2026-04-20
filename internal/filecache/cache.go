package filecache

import (
	"net/http"
	"os"
	"time"

	lru "github.com/jellydator/ttlcache/v3"
)

type Entry struct {
	Content     []byte
	ContentType string
}

type Cache struct {
	inner *lru.Cache[string, *Entry]
}

func New(ttl time.Duration) *Cache {
	c := &Cache{inner: lru.New(lru.WithTTL[string, *Entry](ttl))}

	go c.inner.Start()

	return c
}

func (c *Cache) Get(path string, detectContentType bool) (*Entry, error) {
	if item := c.inner.Get(path); item != nil {
		return item.Value(), nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // path sanitized upstream
	if err != nil {
		return nil, err
	}

	entry := &Entry{Content: data}
	if detectContentType {
		entry.ContentType = http.DetectContentType(data)
	}

	c.inner.Set(path, entry, lru.DefaultTTL)

	return entry, nil
}

func (c *Cache) Stop() {
	c.inner.Stop()
}
