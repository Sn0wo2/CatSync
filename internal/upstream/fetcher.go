package upstream

import (
	"net/http"
	"time"

	"github.com/Sn0wo2/CatSync/internal/cache"
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
	cache *cache.TTL[string, *Response]
}

func NewFetcher(ttl time.Duration) *Fetcher {
	return &Fetcher{cache: cache.NewTTL[string, *Response](ttl)}
}

func (f *Fetcher) Fetch(url string) (*Response, error) {
	if resp, ok := f.cache.Get(url); ok {
		return resp, nil
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

	f.cache.Set(url, r)

	return r, nil
}

func (f *Fetcher) Stop() {
	f.cache.Stop()
}
