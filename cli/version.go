package cli

import (
	"fmt"
	"time"
)

var (
	version = "0.0.0"
	commit  = "dev"
	date    = "1970-01-01T00:00:00Z"
	osName  = "unknown"
	arch    = "unknown"
)

func GetVersion() string {
	return version
}

func SetVersion(ver string) {
	version = ver
}

func GetFormatVersion() string {
	return fmt.Sprintf("v%s-%s(%s)", GetVersion(), GetShortCommit(), GetDateLocal().Format("20060102"))
}

func GetCLIVersion() string {
	return fmt.Sprintf("CatSync[%s]: \ncommit: %s\nbuild date: %s\nos: %s\narch: %s",
		GetFormatVersion(), GetCommit(), GetDateLocal(), GetOSName(), GetArch())
}

func GetCommit() string {
	return commit
}

func GetShortCommit() string {
	if len(commit) > 7 {
		return commit[:7]
	}

	return commit
}

func SetCommit(comm string) {
	commit = comm
}

func GetDate() string {
	return date
}

func GetDateTime() time.Time {
	t, _ := time.Parse(time.RFC3339, date)

	return t
}

func GetDateLocal() time.Time {
	return GetDateTime().Local() //nolint:gosmopolitan
}

func SetDate(dat string) {
	date = dat
}

func GetOSName() string {
	return osName
}

func SetOSName(o string) {
	osName = o
}

func GetArch() string {
	return arch
}

func SetArch(a string) {
	arch = a
}
