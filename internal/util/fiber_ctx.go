package util

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sn0wo2/go-common/helper"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
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

type lazyCtxStringer struct {
	ctx fiber.Ctx
}

func (l lazyCtxStringer) String() string {
	return FiberContextString(l.ctx)
}

// when the log entry is actually written (i.e., log level is enabled).
func LazyFiberContext(ctx fiber.Ctx) zap.Field {
	return zap.Stringer("ctx", lazyCtxStringer{ctx: ctx})
}
