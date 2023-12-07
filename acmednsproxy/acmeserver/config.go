package acmeserver

import (
	"os"

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
	r, err := os.Open(path)
	if err != nil {
		return
	}
	defer r.Close()

	if err = yaml.NewDecoder(r).Decode(&config); err != nil {
		return
	}

	if config.Listen == "" {
		if config.HasTLS() {
			config.Listen = DefaultTLSAddr
		} else {
			config.Listen = DefaultAddr
		}
	}

	return config, nil
}
