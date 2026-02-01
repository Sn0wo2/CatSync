package main

import (
	"context"

	"github.com/Sn0wo2/CatSync/config/reader"
)

func sval(r *reader.String) string {
	if r == nil {
		return ""
	}

	s, _ := r.ReadString(context.Background())

	return s
}
