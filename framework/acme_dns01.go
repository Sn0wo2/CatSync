//go:build catsync_all || feature_acme_dns01

package framework

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"go.uber.org/zap"
)

type providerTimeoutWrapper struct {
	challenge.Provider
	timeout  time.Duration
	interval time.Duration
}

func (p providerTimeoutWrapper) Timeout() (timeout, interval time.Duration) {
	return p.timeout, p.interval
}

type legoUser struct {
	email string
	reg   *registration.Resource
	key   crypto.PrivateKey
}

func (u *legoUser) GetEmail() string {
	return u.email
}

func (u *legoUser) GetRegistration() *registration.Resource {
	return u.reg
}

func (u *legoUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type execDNSProvider struct {
	presentCmd []string
	cleanupCmd []string

	propTimeout  time.Duration
	pollInterval time.Duration
}

func newExecDNSProvider(cfg *config.ServerACMEDNS01) (*execDNSProvider, error) {
	if cfg == nil {
		return nil, errors.New("nil dns config")
	}

	if len(cfg.PresentCmd) == 0 {
		return nil, errors.New("dns.presentCmd is empty")
	}
	if len(cfg.CleanUpCmd) == 0 {
		return nil, errors.New("dns.cleanupCmd is empty")
	}

	t := time.Duration(cfg.PropagationTimeoutSeconds) * time.Second
	if t <= 0 {
		t = dns01.DefaultPropagationTimeout
	}

	i := time.Duration(cfg.PollingIntervalSeconds) * time.Second
	if i <= 0 {
		i = dns01.DefaultPollingInterval
	}

	return &execDNSProvider{
		presentCmd:   cfg.PresentCmd,
		cleanupCmd:   cfg.CleanUpCmd,
		propTimeout:  t,
		pollInterval: i,
	}, nil
}

func (p *execDNSProvider) Timeout() (timeout, interval time.Duration) {
	return p.propTimeout, p.pollInterval
}

func (p *execDNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	return runDNSCmd(p.presentCmd, domain, info.FQDN, info.Value, token, keyAuth)
}

func (p *execDNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	return runDNSCmd(p.cleanupCmd, domain, info.FQDN, info.Value, token, keyAuth)
}

func runDNSCmd(cmd []string, domain, fqdn, value, token, keyAuth string) error {
	if len(cmd) == 0 {
		return errors.New("empty command")
	}

	repl := func(s string) string {
		s = strings.ReplaceAll(s, "{DOMAIN}", domain)
		s = strings.ReplaceAll(s, "{FQDN}", fqdn)
		s = strings.ReplaceAll(s, "{VALUE}", value)
		s = strings.ReplaceAll(s, "{TOKEN}", token)
		s = strings.ReplaceAll(s, "{KEYAUTH}", keyAuth)
		return s
	}

	argv := make([]string, 0, len(cmd))
	for _, s := range cmd {
		argv = append(argv, repl(s))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, argv[0], argv[1:]...)
	output, err := c.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out == "" {
			return fmt.Errorf("dns cmd failed: %w", err)
		}
		return fmt.Errorf("dns cmd failed: %w: %s", err, out)
	}

	return nil
}

func startACMEDNS01(server tlsConfigSetter, acmeCfg *config.ServerACME, logger *zap.Logger) error {
	if acmeCfg == nil {
		return errors.New("nil acme config")
	}
	if logger == nil {
		return errors.New("nil logger")
	}
	if len(acmeCfg.Hosts) == 0 {
		return errors.New("acme.hosts is empty")
	}
	if acmeCfg.DNS01 == nil {
		return errors.New("acme.dns01 is required")
	}

	cacheDir := reader.Trim(acmeCfg.CacheDir)
	if cacheDir == "" {
		cacheDir = "./data/acme"
	}

	certDir := filepath.Join(cacheDir, "dns01")
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return err
	}

	accountKeyPath := filepath.Join(certDir, "account.key")
	accountKey, err := loadOrCreatePrivateKey(accountKeyPath, certcrypto.EC256)
	if err != nil {
		return err
	}

	user := &legoUser{email: reader.Trim(acmeCfg.Email), key: accountKey}
	legoCfg := lego.NewConfig(user)
	legoCfg.CADirURL = lego.LEDirectoryProduction
	dirURL := reader.Trim(acmeCfg.DirectoryURL)
	if dirURL != "" {
		legoCfg.CADirURL = dirURL
	}
	legoCfg.UserAgent = "CatSync"
	legoCfg.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(legoCfg)
	if err != nil {
		return err
	}

	providerName := strings.ToLower(reader.Trim(acmeCfg.DNS01.Provider))
	if providerName == "" {
		providerName = "exec"
	}

	var provider challenge.Provider
	if providerName == "exec" {
		execProvider, err := newExecDNSProvider(acmeCfg.DNS01)
		if err != nil {
			return err
		}
		provider = execProvider
	} else {
		p, err := newDNSProviderByName(providerName)
		if err != nil {
			return err
		}
		provider = p
	}

	// Propagation settings.
	propTimeout := time.Duration(acmeCfg.DNS01.PropagationTimeoutSeconds) * time.Second
	if propTimeout <= 0 {
		propTimeout = dns01.DefaultPropagationTimeout
	}
	pollInterval := time.Duration(acmeCfg.DNS01.PollingIntervalSeconds) * time.Second
	if pollInterval <= 0 {
		pollInterval = dns01.DefaultPollingInterval
	}

	provider = providerTimeoutWrapper{Provider: provider, timeout: propTimeout, interval: pollInterval}

	if err := client.Challenge.SetDNS01Provider(provider); err != nil {
		return err
	}

	if reg, err := client.Registration.ResolveAccountByKey(); err == nil {
		user.reg = reg
	} else {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		user.reg = reg
	}

	mainHost := sanitizeFilename(acmeCfg.Hosts[0])
	certKeyPath := filepath.Join(certDir, mainHost+".key")
	certPath := filepath.Join(certDir, mainHost+".crt")

	certKey, err := loadOrCreatePrivateKey(certKeyPath, certcrypto.RSA2048)
	if err != nil {
		return err
	}

	getAndStore := func() (*tls.Certificate, error) {
		res, err := client.Certificate.Obtain(certificate.ObtainRequest{
			Domains:    acmeCfg.Hosts,
			PrivateKey: certKey,
			Bundle:     true,
		})
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(certPath, res.Certificate, 0o644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(certKeyPath, res.PrivateKey, 0o600); err != nil {
			return nil, err
		}

		pair, err := tls.X509KeyPair(res.Certificate, res.PrivateKey)
		if err != nil {
			return nil, err
		}

		pair.Leaf = nil
		return &pair, nil
	}

	loadFromDisk := func() (*tls.Certificate, error) {
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		keyPEM, err := os.ReadFile(certKeyPath)
		if err != nil {
			return nil, err
		}
		pair, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, err
		}
		pair.Leaf = nil
		return &pair, nil
	}

	var current atomic.Value
	current.Store((*tls.Certificate)(nil))

	setCurrent := func(c *tls.Certificate) {
		current.Store(c)
	}

	cert, err := loadFromDisk()
	if err != nil {
		logger.Info("DNS-01 certificate not found; obtaining", zap.Strings("hosts", acmeCfg.Hosts))
		cert, err = getAndStore()
		if err != nil {
			return err
		}
	}
	setCurrent(cert)

	server.SetTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			c := current.Load().(*tls.Certificate)
			if c == nil {
				return nil, errors.New("no certificate loaded")
			}
			return c, nil
		},
	})

	// Renew loop: re-obtain when expiring soon.
	go func() {
		const (
			checkEvery  = 12 * time.Hour
			renewBefore = 30 * 24 * time.Hour
		)

		for {
			time.Sleep(checkEvery)

			c := current.Load().(*tls.Certificate)
			if c == nil {
				continue
			}

			notAfter, err := getCertNotAfter(c)
			if err != nil {
				logger.Warn("DNS-01 cert parse failed", zap.Error(err))
				continue
			}

			if time.Until(notAfter) > renewBefore {
				continue
			}

			logger.Info("DNS-01 certificate renewing", zap.Time("notAfter", notAfter), zap.Strings("hosts", acmeCfg.Hosts))
			newCert, err := getAndStore()
			if err != nil {
				logger.Warn("DNS-01 certificate renew failed", zap.Error(err))
				continue
			}
			setCurrent(newCert)
			logger.Info("DNS-01 certificate renewed", zap.Time("notAfter", notAfter))
		}
	}()

	logger.Info("ACME DNS-01 enabled", zap.Strings("hosts", acmeCfg.Hosts), zap.String("cacheDir", certDir))
	return nil
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "*", "_wildcard_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

func loadOrCreatePrivateKey(path string, keyType certcrypto.KeyType) (crypto.PrivateKey, error) {
	if b, err := os.ReadFile(path); err == nil {
		k, err := certcrypto.ParsePEMPrivateKey(b)
		if err == nil {
			return k, nil
		}
	}

	k, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, certcrypto.PEMEncode(k), 0o600); err != nil {
		return nil, err
	}

	return k, nil
}

func getCertNotAfter(c *tls.Certificate) (time.Time, error) {
	if c.Leaf != nil {
		return c.Leaf.NotAfter, nil
	}
	if len(c.Certificate) == 0 {
		return time.Time{}, errors.New("empty certificate")
	}
	leaf, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return time.Time{}, err
	}
	return leaf.NotAfter, nil
}
