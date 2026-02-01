package framework

import "crypto/tls"

type tlsConfigSetter interface {
	SetTLSConfig(cfg *tls.Config)
}
