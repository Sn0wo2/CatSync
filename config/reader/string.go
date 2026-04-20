package reader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type StringType string

const (
	StringTypeAuto   StringType = "auto"
	StringTypeString StringType = "string"
	StringTypePath   StringType = "path"
	StringTypeHTTP   StringType = "http"
)

type String struct {
	Type    StringType `json:"type,omitempty" optional:"true" yaml:"type,omitempty"`
	Content string     `json:"content"        yaml:"content"`
	loaded  bool
	value   string
	err     error
	mu      sync.RWMutex
}

func Str(s string) *String {
	return &String{Type: StringTypeString, Content: s}
}

func (r *String) Get() (string, error) {
	if r == nil {
		return "", nil
	}

	return r.ReadString(context.Background())
}

func (r *String) Must() string {
	s, _ := r.Get()

	return s
}

func (r *String) Trim() string {
	return strings.TrimSpace(r.Must())
}

func (r *String) Literal() (string, bool) {
	if r == nil {
		return "", true
	}

	if r.Type == "" || r.Type == StringTypeString {
		return r.Content, true
	}

	return "", false
}

func (r *String) LiteralTrim() (string, bool) {
	s, ok := r.Literal()

	return strings.TrimSpace(s), ok
}

func (r *String) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		r.Type = StringTypeString
		r.Content = s

		return nil
	}

	type alias String

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	r.Type = a.Type

	r.Content = a.Content
	if r.Type == "" {
		r.Type = StringTypeString
	}

	return nil
}

func (r *String) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}

	if r.Type == "" || r.Type == StringTypeString {
		return json.Marshal(r.Content)
	}

	type alias String

	return json.Marshal((*alias)(r))
}

func (r *String) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.ScalarNode {
		r.Type = StringTypeString
		r.Content = node.Value

		return nil
	}

	type alias String

	var a alias
	if err := node.Decode(&a); err != nil {
		return err
	}

	r.Type = a.Type

	r.Content = a.Content
	if r.Type == "" {
		r.Type = StringTypeString
	}

	return nil
}

func (r *String) MarshalYAML() (any, error) {
	if r == nil {
		return nil, nil
	}

	if r.Type == "" || r.Type == StringTypeString {
		return r.Content, nil
	}

	type alias String

	return (*alias)(r), nil
}

func (r *String) ValidateNoIO() error {
	if r == nil {
		return nil
	}

	if strings.TrimSpace(r.Content) == "" {
		return errors.New("content is empty")
	}

	switch r.Type {
	case "", StringTypeString, StringTypeAuto:
		return nil
	case StringTypePath:
		return nil
	case StringTypeHTTP:
		_, err := parseHTTPURL(r.Content)

		return err
	default:
		return fmt.Errorf("invalid type: %q", r.Type)
	}
}

func (r *String) ReadString(ctx context.Context) (string, error) {
	if r == nil {
		return "", nil
	}

	r.mu.RLock()

	if r.loaded {
		r.mu.RUnlock()

		return r.value, r.err
	}

	r.mu.RUnlock()
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return r.value, r.err
	}

	t, content := resolve(r.Type, r.Content)
	switch t {
	case StringTypeAuto:
		r.value = content
	case StringTypeString:
		r.value = content
	case StringTypePath:
		b, err := os.ReadFile(content)
		if err != nil {
			r.err = err
			r.loaded = true

			return r.value, r.err
		}

		r.value = string(b)
	case StringTypeHTTP:
		u, err := parseHTTPURL(content)
		if err != nil {
			r.err = err
			r.loaded = true

			return r.value, r.err
		}

		c := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			r.err = err
			r.loaded = true

			return r.value, r.err
		}

		resp, err := c.Do(req)
		if err != nil {
			r.err = err
			r.loaded = true

			return r.value, r.err
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			r.err = fmt.Errorf("http status %d", resp.StatusCode)
			r.loaded = true

			return r.value, r.err
		}

		b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			r.err = err
			r.loaded = true

			return r.value, r.err
		}

		r.value = string(b)
	default:
		r.err = fmt.Errorf("unsupported type: %q", t)
	}

	r.loaded = true

	return r.value, r.err
}

func (r *String) Reset() {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.loaded = false
	r.value = ""
	r.err = nil
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
