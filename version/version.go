package version

import (
	"fmt"
	"time"
)

var (
	version = "0.0.0"
	commit  = "dev"
	date    = "1970-01-01T00:00:00Z"
	os      = "unknown"
	arch    = "unknown"
)

// GetVersion without 'v'
func GetVersion() string {
	return version
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

func GetDate() string {
	return date
}

func GetDateTime() time.Time {
	t, _ := time.Parse(time.RFC3339, date)
	return t
}

func GetFormatVersion() string {
	return fmt.Sprintf("v%s-%s(%s)", GetVersion(), GetShortCommit(), GetDate())
}

func GetCLIVersion() string {
	return fmt.Sprintf("version: %s\ncommit: %s\nbuild date: UTC+8 %s\nos: %s\narch: %s", GetVersion(), GetCommit(), GetDate(), GetOS(), GetArch())
}

// SetVersion without 'v'
func SetVersion(ver string) {
	version = ver
}

func SetCommit(comm string) {
	commit = comm
}

func SetDate(dat string) {
	date = dat
}

func GetOS() string {
	return os
}

func SetOS(o string) {
	os = o
}

func GetArch() string {
	return arch
}

func SetArch(a string) {
	arch = a
}
