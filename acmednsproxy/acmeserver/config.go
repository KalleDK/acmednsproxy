// Package acmeserver wires together the HTTP server, TLS certificate management,
// and the acmeservice.DNSProxy.  It exposes a Server and a ServerWithConfig
// (which supports hot-reload) that together handle the full lifecycle of the
// acmednsproxy service.
package acmeserver

import (
	"os"
	"path/filepath"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"gopkg.in/yaml.v3"
)

// DefaultAddr is the listen address used when no TLS is configured.
const DefaultAddr = ":8080"

// DefaultTLSAddr is the listen address used when TLS is configured.
const DefaultTLSAddr = ":9090"

// Config holds the top-level server configuration decoded from the main YAML
// file (typically acmednsproxy.yaml).
type Config struct {
	// Listen is the TCP address the server binds to (e.g. ":9090").
	// If empty, it defaults to DefaultTLSAddr when TLS is configured or
	// DefaultAddr otherwise.
	Listen string
	// TLS holds the paths to the TLS certificate and key files.
	TLS TLSConfig
	// Proxy holds the paths to the authenticator and provider config files.
	Proxy acmeservice.Config `yaml:",inline"`
}

// HasTLS reports whether TLS has been configured.
func (c Config) HasTLS() bool {
	return !c.TLS.IsEmpty()
}

// loadConfig reads and parses the YAML config file at path.  Relative paths
// inside the config are resolved relative to the directory containing path.
func loadConfig(path string) (config Config, err error) {
	confDir := filepath.Dir(path)

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
			config.TLS.CertFile = filepath.Join(confDir, config.TLS.CertFile)
		}

		if !filepath.IsAbs(config.TLS.KeyFile) {
			config.TLS.KeyFile = filepath.Join(confDir, config.TLS.KeyFile)
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
		config.Proxy.Provider = filepath.Join(confDir, config.Proxy.Provider)
	}

	if !filepath.IsAbs(config.Proxy.Authenticator) {
		config.Proxy.Authenticator = filepath.Join(confDir, config.Proxy.Authenticator)
	}

	return config, nil
}
