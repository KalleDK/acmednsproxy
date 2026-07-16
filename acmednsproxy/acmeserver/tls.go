package acmeserver

import (
	"crypto/tls"
	"fmt"
)

// TLSConfig holds the file-system paths to the TLS certificate and private key.
type TLSConfig struct {
	// CertFile is the path to the PEM-encoded TLS certificate (chain).
	CertFile string `yaml:",omitempty"`
	// KeyFile is the path to the PEM-encoded private key.
	KeyFile string `yaml:",omitempty"`
}

// IsEmpty reports whether neither CertFile nor KeyFile has been set.
func (c TLSConfig) IsEmpty() bool {
	return c.CertFile == "" && c.KeyFile == ""
}

// loadCert reads the certificate and key files specified in config and returns
// a parsed tls.Certificate.
func loadCert(config TLSConfig) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
}

// TLSService holds a parsed TLS certificate and can reload it from disk
// without restarting the server.  The GetCertificate method is passed directly
// to tls.Config so that the in-memory certificate is refreshed on every reload.
type TLSService struct {
	Config      TLSConfig
	Certificate tls.Certificate
}

// GetCertificate implements the tls.Config.GetCertificate callback.  It
// returns the most recently loaded certificate so that certificate rotations
// (triggered via the /reload endpoint) take effect without a server restart.
func (c *TLSService) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if c == nil {
		return nil, fmt.Errorf("not configured for tls")
	}
	return &c.Certificate, nil
}

// Reload re-reads the certificate and key from disk and replaces the
// in-memory certificate.  It is safe to call on a nil receiver (no-op).
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

// NewTLSService loads the certificate and key from config and returns a
// TLSService ready for use.
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
