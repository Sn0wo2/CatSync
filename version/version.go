package version

import (
	"fmt"
	"time"
)

var (
	// version without 'v'
	version = "0.0.0"
	commit  = "dev"
	// date RFC3339 UTC
	date = "1970-01-01T00:00:00Z"
	os   = "unknown"
	arch = "unknown"
)

// GetVersion without 'v'
func GetVersion() string {
	return version
}

func GetFormatVersion() string {
	return fmt.Sprintf("v%s-%s(%s)", GetVersion(), GetShortCommit(), GetDateLocal())
}

func GetCLIVersion() string {
	return fmt.Sprintf("version: %s\ncommit: %s\nbuild date: %s\nos: %s\narch: %s",
		GetVersion(), GetCommit(), GetDateLocal(), GetOS(), GetArch())
}

// SetVersion without 'v'
func SetVersion(ver string) {
	version = ver
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

// GetDate RFC3339 UTC
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

// SetDate RFC3339 UTC
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
