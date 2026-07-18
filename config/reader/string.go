package reader

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
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
