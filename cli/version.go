package cli

import (
	"cmp"
	"fmt"
	"time"

	"github.com/prometheus/common/version"
)

func GetFormatVersion() string {
	commit := cmp.Or(version.Revision, "dev")
	t, _ := time.Parse(time.RFC3339, cmp.Or(version.BuildDate, "1970-01-01T00:00:00Z"))

	return fmt.Sprintf("v%s-%s(%s)", cmp.Or(version.Version, "0.0.0"), commit[:min(7, len(commit))], t.Local().Format("20060102")) //nolint:gosmopolitan
}
