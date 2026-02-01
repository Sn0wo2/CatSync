//go:build catsync_all || feature_acme_http01

package framework

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func startACMEHTTP01(server *http.Server, cfg *config.Config, acmeCfg *config.ServerACME, logger *zap.Logger) error {
	if server == nil {
		return fmt.Errorf("nil http server")
	}
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	if acmeCfg == nil {
		return fmt.Errorf("nil acme config")
	}
	if logger == nil {
		return fmt.Errorf("nil logger")
	}

	cacheDir := strings.TrimSpace(acmeCfg.CacheDir)
	if cacheDir == "" {
		cacheDir = "./data/acme"
	}

	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cacheDir),
		Email:  strings.TrimSpace(acmeCfg.Email),
	}

	if len(acmeCfg.Hosts) > 0 {
		m.HostPolicy = autocert.HostWhitelist(acmeCfg.Hosts...)
	}
	if strings.TrimSpace(acmeCfg.DirectoryURL) != "" {
		m.Client = &acme.Client{DirectoryURL: strings.TrimSpace(acmeCfg.DirectoryURL)}
	}

	server.TLSConfig = m.TLSConfig()

	httpAddr := strings.TrimSpace(acmeCfg.HTTPAddress)
	if httpAddr == "" {
		httpAddr = ":80"
	}

	redirect := true
	if cfg.Server.TLS.RedirectHTTP != nil {
		redirect = *cfg.Server.TLS.RedirectHTTP
	}

	tlsPort := "443"
	if _, port, err := net.SplitHostPort(cfg.Server.Address); err == nil && port != "" {
		tlsPort = port
	}

	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !redirect {
			http.NotFound(w, r)
			return
		}

		host := r.Host
		// If request host has a port, strip it.
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if tlsPort != "443" {
			host = fmt.Sprintf("%s:%s", host, tlsPort)
		}

		target := "https://" + host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})

	// Start HTTP-01 challenge handler.
	go func() {
		logger.Info("ACME HTTP challenge listening", zap.String("addr", httpAddr))
		h := m.HTTPHandler(redirectHandler)
		if err := http.ListenAndServe(httpAddr, h); err != nil {
			logger.Warn("ACME HTTP challenge server stopped", zap.Error(err))
		}
	}()

	logger.Info("ACME enabled (http-01)", zap.Strings("hosts", acmeCfg.Hosts), zap.String("cacheDir", cacheDir))
	return server.ListenAndServeTLS("", "")
}
