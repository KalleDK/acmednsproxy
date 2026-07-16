// Package acmeservice wires together an Authenticator and a DNSProviders
// multiplexer into a DNSProxy that exposes the high-level Present / Cleanup /
// Authenticate operations used by the HTTP layer.
package acmeservice

import (
	"context"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

// Config holds the file-system paths to the two YAML configuration files that
// the DNSProxy loads at startup (and on every Reload).
type Config struct {
	// Authenticator is the path to the auth YAML file (e.g. auth.yaml).
	Authenticator string
	// Provider is the path to the providers YAML file (e.g. provider.yaml).
	Provider string
}

// DNSProxy combines an Authenticator and a DNSProviders multiplexer.  It is
// the central service object used by the HTTP handlers.
type DNSProxy struct {
	Config   Config
	Auth     auth.Authenticator
	Provider *providers.DNSProviders
}

// Shutdown gracefully shuts down the authenticator and all DNS providers.
func (s *DNSProxy) Shutdown(ctx context.Context) error {
	if err := s.Auth.Shutdown(ctx); err != nil {
		return err
	}

	if err := s.Provider.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// Close immediately closes the authenticator and all DNS providers.
func (s *DNSProxy) Close() error {
	if err := s.Auth.Close(); err != nil {
		return err
	}

	if err := s.Provider.Close(); err != nil {
		return err
	}

	return nil
}

// Reload atomically replaces the authenticator and provider set by loading
// fresh copies from the config files.  The old instances are closed after the
// swap so in-flight requests can still complete.
func (s *DNSProxy) Reload() (err error) {
	var oldAuth, newAuth auth.Authenticator
	var oldProv, newProv *providers.DNSProviders

	if newAuth, err = auth.LoadFromFile(s.Config.Authenticator); err != nil {
		return err
	}

	if newProv, err = providers.LoadFromFile(s.Config.Provider); err != nil {
		return err
	}

	oldAuth, s.Auth = s.Auth, newAuth
	oldProv, s.Provider = s.Provider, newProv

	if oldAuth != nil {
		oldAuth.Close()
	}

	if oldProv != nil {
		oldProv.Close()
	}

	return nil
}

// Authenticate verifies that cred is authorised to manage DNS records for
// domain.  It delegates to the underlying Authenticator.
func (s *DNSProxy) Authenticate(cred auth.Credentials, domain string) error {
	return s.Auth.VerifyPermissions(cred, domain)
}

// Present publishes a DNS TXT challenge record via the appropriate provider.
func (s *DNSProxy) Present(record providers.Record) error {
	return s.Provider.CreateRecord(record)
}

// Cleanup removes a previously published DNS TXT challenge record.
func (s *DNSProxy) Cleanup(record providers.Record) error {
	return s.Provider.RemoveRecord(record)
}

// New constructs a DNSProxy by loading the authenticator and provider files
// referenced by config.
func New(config Config) (proxy *DNSProxy, err error) {
	var a auth.Authenticator
	var prov *providers.DNSProviders

	if a, err = auth.LoadFromFile(config.Authenticator); err != nil {
		return
	}

	if prov, err = providers.LoadFromFile(config.Provider); err != nil {
		return
	}

	return &DNSProxy{
		Config:   config,
		Auth:     a,
		Provider: prov,
	}, nil
}
