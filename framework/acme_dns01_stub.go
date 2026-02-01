//go:build !catsync_all && !feature_acme_dns01

package framework

import (
	"fmt"
	"net/http"

	"github.com/Sn0wo2/CatSync/config"
	"go.uber.org/zap"
)

func startACMEDNS01(_ *http.Server, _ *config.ServerACME, _ *zap.Logger) error {
	return fmt.Errorf("ACME dns-01 is not enabled in this build; rebuild with -tags catsync_all or -tags feature_acme_dns01")
}
