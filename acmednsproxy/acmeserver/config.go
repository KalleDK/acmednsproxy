package acmeserver

import (
	"os"
	"path/filepath"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"gopkg.in/yaml.v3"
)

const (
	DefaultAddr    = ":8080"
	DefaultTLSAddr = ":9090"
)

type Config struct {
	Listen string
	TLS    TLSConfig
	Proxy  acmeservice.Config `yaml:",inline"`
}

func (c Config) HasTLS() bool {
	return !c.TLS.IsEmpty()
}

func loadConfig(path string) (config Config, err error) {
	conf_dir := filepath.Dir(path)

	r, err := os.Open(path)
	if err != nil {
		return
	}
	defer r.Close()

	if err = yaml.NewDecoder(r).Decode(&config); err != nil {
		return
	}

	if config.HasTLS() {
		if !filepath.IsAbs(config.TLS.CertFile) {
			config.TLS.CertFile = filepath.Join(conf_dir, config.TLS.CertFile)
		}

		if !filepath.IsAbs(config.TLS.KeyFile) {
			config.TLS.KeyFile = filepath.Join(conf_dir, config.TLS.KeyFile)
		}
	}

	if config.Listen == "" {
		if config.HasTLS() {
			config.Listen = DefaultTLSAddr
		} else {
			config.Listen = DefaultAddr
		}
	}

	if !filepath.IsAbs(config.Proxy.Provider) {
		config.Proxy.Provider = filepath.Join(conf_dir, config.Proxy.Provider)
	}

	if !filepath.IsAbs(config.Proxy.Authenticator) {
		config.Proxy.Authenticator = filepath.Join(conf_dir, config.Proxy.Authenticator)
	}

	return config, nil
}
