package util

import (
	"log/slog"
	"runtime/debug"
)

type LazyStack struct{}

func (LazyStack) LogValue() slog.Value {
	return slog.StringValue(string(debug.Stack()))
}
