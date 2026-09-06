package reader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func (r *String) ReadString(ctx context.Context) (string, error) {
	if r == nil {
		return "", nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.once.Do(func() {
		fail := func(err error) {
			r.err = err
		}

		t, content := resolve(r.Type, r.Content)
		switch t {
		case StringTypeAuto, StringTypeString:
			r.value = content
		case StringTypePath:
			//nolint:gosec // This is an intended feature: reading file contents via user configuration.
			b, err := os.ReadFile(content)
			if err != nil {
				fail(err)

				return
			}

			r.value = string(b)
		case StringTypeHTTP:
			u, err := parseHTTPURL(content)
			if err != nil {
				fail(err)

				return
			}

			c := &http.Client{Timeout: 10 * time.Second}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				fail(err)

				return
			}

			resp, err := c.Do(req)
			if err != nil {
				fail(err)

				return
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				fail(fmt.Errorf("http status %d", resp.StatusCode))

				return
			}

			b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			if err != nil {
				fail(err)

				return
			}

			r.value = string(b)
		default:
			fail(fmt.Errorf("unsupported type: %q", t))
		}
	})

	return r.value, r.err
}

func (r *String) Reset() {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.value = ""
	r.err = nil
	r.once = sync.Once{}
}

func (r *String) ReadLines(ctx context.Context) ([]string, error) {
	s, err := r.ReadString(ctx)
	if err != nil {
		return nil, err
	}

	if s == "" {
		return nil, nil
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	return strings.Split(s, "\n"), nil
}

func resolve(t StringType, content string) (StringType, string) {
	if t == "" {
		t = StringTypeString
	}

	if t != StringTypeAuto {
		return t, content
	}

	s := strings.TrimSpace(content)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return StringTypeHTTP, content
	}

	if looksLikePath(s) {
		return StringTypePath, content
	}

	return StringTypeString, content
}

func looksLikePath(s string) bool {
	if s == "" {
		return false
	}

	if strings.ContainsAny(s, "\\/") {
		return true
	}

	if strings.HasPrefix(s, ".") {
		return true
	}

	if len(s) >= 2 && (s[1] == ':' && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z'))) {
		return true
	}

	return false
}

func parseHTTPURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}

	if u == nil || u.Scheme == "" {
		return nil, errors.New("url is empty")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme: %q", u.Scheme)
	}

	if u.Host == "" {
		return nil, errors.New("url host is empty")
	}

	return u, nil
}
