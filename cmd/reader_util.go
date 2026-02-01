package main

import "github.com/Sn0wo2/CatSync/config/reader"

func sval(r *reader.String) string {
	return reader.Must(r)
}
