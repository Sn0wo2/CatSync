//go:build !catsync_all && !feature_acme_http01

package framework

import (
	"errors"

	"github.com/Sn0wo2/CatSync/config"
	"go.uber.org/zap"
)

func startACMEHTTP01(_ tlsConfigSetter, _ *config.Config, _ *config.ServerACME, _ *zap.Logger) error {
	return errors.New("ACME http-01 is not enabled in this build; rebuild with -tags catsync_all or -tags feature_acme_http01")
}
