package acmeservice

import (
	"context"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

type Config struct {
	Authenticator string
	Provider      string
}

type DNSProxy struct {
	Config   Config
	Auth     auth.Authenticator
	Provider providers.DNSProvider
}

func (s *DNSProxy) Shutdown(ctx context.Context) error {
	if err := s.Auth.Shutdown(ctx); err != nil {
		return err
	}

	if err := s.Provider.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (s *DNSProxy) Close() error {
	if err := s.Auth.Close(); err != nil {
		return err
	}

	if err := s.Provider.Close(); err != nil {
		return err
	}

	return nil
}

func (s *DNSProxy) Reload() (err error) {
	var old_auth, new_auth auth.Authenticator
	var old_prov, new_prov providers.DNSProvider

	if new_auth, err = auth.LoadFromFile(s.Config.Authenticator); err != nil {
		return err
	}

	if new_prov, err = providers.LoadFromFile(s.Config.Provider); err != nil {
		return err
	}

	old_auth, s.Auth = s.Auth, new_auth
	old_prov, s.Provider = s.Provider, new_prov

	if old_auth != nil {
		old_auth.Close()
	}

	if old_prov != nil {
		old_prov.Close()
	}

	return nil
}

func (s *DNSProxy) Authenticate(cred auth.Credentials, domain string) error {

	if err := s.Auth.VerifyPermissions(cred, domain); err != nil {
		return err
	}
	return nil
}

func (s *DNSProxy) Present(record providers.Record) error {
	if err := s.Provider.CreateRecord(record); err != nil {
		return err
	}

	return nil
}

func (s *DNSProxy) Cleanup(record providers.Record) error {
	if err := s.Provider.RemoveRecord(record); err != nil {
		return err
	}

	return nil
}

func New(config Config) (proxy *DNSProxy, err error) {
	var a auth.Authenticator
	var prov providers.DNSProvider

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
