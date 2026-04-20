package filecache

import (
	"net/http"
	"os"
	"time"

	"github.com/Sn0wo2/CatSync/internal/cache"
)

type Entry struct {
	Content     []byte
	ContentType string
}

type Cache struct {
	inner *cache.TTL[string, *Entry]
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		inner: cache.NewTTL[string, *Entry](ttl),
	}

	return c
}

func (c *Cache) Get(path string, detectContentType bool) (*Entry, error) {
	if entry, ok := c.inner.Get(path); ok {
		return entry, nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // path sanitized upstream
	if err != nil {
		return nil, err
	}

	entry := &Entry{Content: data}
	if detectContentType {
		entry.ContentType = http.DetectContentType(data)
	}

	c.inner.Set(path, entry)

	return entry, nil
}

func (c *Cache) Stop() {
	c.inner.Stop()
}
