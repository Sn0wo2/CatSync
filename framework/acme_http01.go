//go:build catsync_all || feature_acme_http01

package framework

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func startACMEHTTP01(server tlsConfigSetter, cfg *config.Config, acmeCfg *config.ServerACME, logger *zap.Logger) error {
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

	cacheDir := reader.Trim(acmeCfg.CacheDir)
	if cacheDir == "" {
		cacheDir = "./data/acme"
	}

	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cacheDir),
		Email:  reader.Trim(acmeCfg.Email),
	}

	if len(acmeCfg.Hosts) > 0 {
		m.HostPolicy = autocert.HostWhitelist(acmeCfg.Hosts...)
	}
	dirURL := reader.Trim(acmeCfg.DirectoryURL)
	if dirURL != "" {
		m.Client = &acme.Client{DirectoryURL: dirURL}
	}

	server.SetTLSConfig(m.TLSConfig())

	httpAddr := ""
	if acmeCfg.HTTP01 != nil {
		httpAddr = reader.Trim(acmeCfg.HTTP01.HTTPAddress)
	}
	if httpAddr == "" {
		httpAddr = ":80"
	}

	redirect := true
	if cfg.Server.TLS.RedirectHTTP != nil {
		redirect = *cfg.Server.TLS.RedirectHTTP
	}

	tlsPort := "443"
	if _, port, err := net.SplitHostPort(reader.Trim(cfg.Server.Address)); err == nil && port != "" {
		tlsPort = port
	}

	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !redirect {
			http.NotFound(w, r)
			return
		}

		// Redirect only to known hosts to avoid host-header injection / open redirects.
		// Note: server.acme.hosts is required when ACME is enabled.
		requestHost := r.Host
		if h, _, err := net.SplitHostPort(requestHost); err == nil {
			requestHost = h
		}

		redirectHost := ""
		for _, allowed := range acmeCfg.Hosts {
			if strings.EqualFold(strings.TrimSpace(allowed), strings.TrimSpace(requestHost)) {
				redirectHost = allowed
				break
			}
		}
		if redirectHost == "" {
			http.Error(w, "invalid host", http.StatusBadRequest)
			return
		}
		if tlsPort != "443" {
			redirectHost = fmt.Sprintf("%s:%s", redirectHost, tlsPort)
		}

		p := r.URL.EscapedPath()
		if p == "" {
			p = "/"
		}
		u := url.URL{Scheme: "https", Host: redirectHost, Path: p, RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
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
	return nil
}
