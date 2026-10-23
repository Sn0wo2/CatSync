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
	"time"
)

func (r *String) ReadString(ctx context.Context) (string, error) {
	if r == nil {
		return "", nil
	}

	r.once.Do(func() {
		t, content := resolve(r.Type, r.Content)
		switch t {
		case StringTypeAuto, StringTypeString:
			r.value = content
		case StringTypePath:
			//nolint:gosec // This is an intended feature: reading file contents via user configuration.
			b, err := os.ReadFile(content)
			if err != nil {
				r.err = err

				return
			}

			r.value = string(b)
		case StringTypeHTTP:
			u, err := parseHTTPURL(content)
			if err != nil {
				r.err = err

				return
			}

			c := &http.Client{Timeout: 10 * time.Second}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				r.err = err

				return
			}

			resp, err := c.Do(req)
			if err != nil {
				r.err = err

				return
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				r.err = fmt.Errorf("http status %d", resp.StatusCode)

				return
			}

			b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			if err != nil {
				r.err = err

				return
			}

			r.value = string(b)
		default:
			r.err = fmt.Errorf("unsupported type: %q", t)
		}
	})

	return r.value, r.err
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

	if strings.ContainsAny(s, "\\/") || strings.HasPrefix(s, ".") ||
		(len(s) >= 2 && s[1] == ':' && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z'))) {
		return StringTypePath, content
	}

	return StringTypeString, content
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
