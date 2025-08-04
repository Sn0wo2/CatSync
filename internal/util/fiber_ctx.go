package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sn0wo2/FileSync/internal/helper"
	"github.com/gofiber/fiber/v2"
)

func FiberContextString(ctx *fiber.Ctx) string {
	var sb strings.Builder

	ips := ctx.IPs()
	if len(ips) == 0 {
		ips = []string{ctx.IP()}
	}

	sb.WriteString(strings.Join(ips, ", "))

	sb.WriteString(" -> ")
	sb.WriteString(ctx.Method())

	sb.WriteString(" ")

	if ctx.Response().StatusCode() != 0 && ctx.Response().StatusCode() != 100 {
		sb.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
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
