package util

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sn0wo2/go-common/helper"
	"github.com/gofiber/fiber/v3"
)

func FiberContextString(ctx fiber.Ctx) string {
	var sb strings.Builder

	ips := ctx.IPs()
	if len(ips) == 0 {
		ips = []string{ctx.IP()}
	}

	sb.WriteString(strings.Join(ips, ", "))

	sb.WriteString(" -> ")
	sb.WriteString(ctx.Method())

	sb.WriteString(" ")

	if ctx.Response().StatusCode() != 0 {
		statusCode := ctx.Response().StatusCode()
		sb.WriteString(strconv.Itoa(statusCode))
		sb.WriteString(" ")
		sb.WriteString(http.StatusText(statusCode))
		sb.WriteString(" ")
	}

	sb.WriteString(helper.BytesToString(ctx.Request().RequestURI()))

	var headers []string

	ctx.Request().Header.All()(func(key, value []byte) bool {
		v := helper.BytesToString(value)
		if len(v) > 20 {
			v = v[:12] + "..."
		}

		headers = append(headers, fmt.Sprintf("%s:%s", helper.BytesToString(key), v))

		return true
	})

	if len(headers) > 0 {
		sb.WriteString(" { ")
		sb.WriteString(strings.Join(headers, ", "))
		sb.WriteString(" }")
	}

	return sb.String()
}

type lazyCtxValue struct {
	ctx fiber.Ctx
}

func (l lazyCtxValue) LogValue() slog.Value {
	return slog.StringValue(FiberContextString(l.ctx))
}

// LazyFiberContext formats the request only when the log entry is written.
func LazyFiberContext(ctx fiber.Ctx) slog.LogValuer {
	return lazyCtxValue{ctx: ctx}
}
