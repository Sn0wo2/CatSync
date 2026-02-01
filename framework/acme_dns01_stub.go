//go:build !catsync_all && !feature_acme_dns01

package framework

import (
	"errors"

	"github.com/Sn0wo2/CatSync/config"
	"go.uber.org/zap"
)

func startACMEDNS01(_ tlsConfigSetter, _ *config.ServerACME, _ *zap.Logger) error {
	return errors.New("ACME dns-01 is not enabled in this build; rebuild with -tags catsync_all or -tags feature_acme_dns01")
}
