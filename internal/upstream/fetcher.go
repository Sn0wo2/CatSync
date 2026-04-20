package upstream

import (
	"net/http"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

type Response struct {
	StatusCode int
	Header     http.Header
}

type Fetcher struct {
	cache *ttlcache.Cache[string, *Response]
}

func NewFetcher(ttl time.Duration) *Fetcher {
	c := ttlcache.New(ttlcache.WithTTL[string, *Response](ttl))
	go c.Start()

	return &Fetcher{cache: c}
}

func (f *Fetcher) Fetch(url string) (*Response, error) {
	if item := f.cache.Get(url); item != nil {
		return item.Value(), nil
	}

	resp, err := sharedClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	r := &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
	}

	f.cache.Set(url, r, ttlcache.DefaultTTL)

	return r, nil
}

func (f *Fetcher) Stop() {
	f.cache.Stop()
}
