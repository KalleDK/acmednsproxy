package acmeserver

import (
	"crypto/tls"
	"fmt"
)

type TLSConfig struct {
	CertFile string `yaml:",omitempty"`
	KeyFile  string `yaml:",omitempty"`
}

func (c TLSConfig) IsEmpty() bool {
	return c.CertFile == "" && c.KeyFile == ""
}

func loadCert(config TLSConfig) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
}

type TLSService struct {
	Config      TLSConfig
	Certificate tls.Certificate
}

func (c *TLSService) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if c == nil {
		return nil, fmt.Errorf("not configured for tls")
	}
	return &c.Certificate, nil
}

func (c *TLSService) Reload() (err error) {
	if c == nil {
		return nil
	}

	cert, err := loadCert(c.Config)
	if err != nil {
		return fmt.Errorf("failed to load tls certificate %+v: %w", c.Config, err)
	}

	c.Certificate = cert
	return nil
}

func NewTLSService(config TLSConfig) (*TLSService, error) {
	cert, err := loadCert(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls certificate %+v: %w", config, err)
	}

	return &TLSService{
		Config:      config,
		Certificate: cert,
	}, nil
}
