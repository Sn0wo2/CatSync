package util

import (
	"log/slog"
	"runtime/debug"
)

type lazyStack struct{}

func (lazyStack) LogValue() slog.Value {
	return slog.StringValue(string(debug.Stack()))
}

func LazyStack() slog.LogValuer {
	return lazyStack{}
}
