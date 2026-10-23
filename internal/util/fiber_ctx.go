package util

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

type LazyFiberContext struct{ fiber.Ctx }

func (ctx LazyFiberContext) LogValue() slog.Value {
	var sb strings.Builder

	ips := ctx.IPs()
	if len(ips) == 0 {
		ips = []string{ctx.IP()}
	}

	sb.WriteString(strings.Join(ips, ", "))

	sb.WriteString(" -> ")
	sb.WriteString(ctx.Method())

	sb.WriteString(" ")

	if statusCode := ctx.Response().StatusCode(); statusCode != 0 {
		sb.WriteString(strconv.Itoa(statusCode))
		sb.WriteString(" ")
		sb.WriteString(http.StatusText(statusCode))
		sb.WriteString(" ")
	}

	sb.WriteString(utils.UnsafeString(ctx.Request().RequestURI()))

	var headers []string

	for key, value := range ctx.Request().Header.All() {
		v := utils.UnsafeString(value)
		if len(v) > 20 {
			v = v[:12] + "..."
		}

		headers = append(headers, utils.UnsafeString(key)+":"+v)
	}

	if len(headers) > 0 {
		sb.WriteString(" { ")
		sb.WriteString(strings.Join(headers, ", "))
		sb.WriteString(" }")
	}

	return slog.StringValue(sb.String())
}
