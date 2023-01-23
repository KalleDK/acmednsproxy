package acmeservice

import (
	"fmt"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

type DNSProxy struct {
	ConfigLoader ConfigLoader
	Auth         auth.Authenticator
	Provider     providers.DNSProvider
}

func (s *DNSProxy) reloadAuth(config AuthConfig, config_dir string) (err error) {
	p, err := config.Load(config_dir)
	if err != nil {
		return err
	}

	if s.Auth != nil {
		if err := s.Auth.Close(); err != nil {
			return err
		}
	}

	s.Auth = p

	return nil
}

func (s *DNSProxy) reloadProvider(config ProviderConfig, config_dir string) (err error) {
	p, err := config.Load(config_dir)
	if err != nil {
		return err
	}

	if s.Provider != nil {
		if err := s.Provider.Close(); err != nil {
			return err
		}
	}

	s.Provider = p

	return nil
}

func (s *DNSProxy) Reload() (err error) {
	config, config_dir, err := s.ConfigLoader.Load()
	fmt.Printf("%+v\n", config)
	if err != nil {
		return err
	}

	if err := s.reloadProvider(config.Provider, config_dir); err != nil {
		return err
	}

	if err := s.reloadAuth(config.Authenticator, config_dir); err != nil {
		return err
	}

	return nil
}

func (s *DNSProxy) Authenticate(auth Auth, record Record) error {
	domain := record.FQDN

	if !strings.HasPrefix(domain, "_acme-challenge.") {
		return fmt.Errorf("invalid challenge domain %s", record.FQDN)
	}
	domain = strings.TrimPrefix(domain, "_acme-challenge.")

	if !strings.HasSuffix(domain, ".") {
		return fmt.Errorf("invalid challenge domain %s", record.FQDN)
	}
	domain = strings.TrimSuffix(domain, ".")

	if err := s.Auth.VerifyPermissions(auth.Username, auth.Password, domain); err != nil {
		return err
	}
	return nil
}

func (s *DNSProxy) Present(record Record) error {
	if err := s.Provider.CreateRecord(record.FQDN, record.Value); err != nil {
		return err
	}

	return nil
}

func (s *DNSProxy) Cleanup(record Record) error {
	if err := s.Provider.RemoveRecord(record.FQDN, record.Value); err != nil {
		return err
	}

	return nil
}

func New(loader ConfigLoader) (*DNSProxy, error) {
	proxy := &DNSProxy{
		ConfigLoader: loader,
	}

	if err := proxy.Reload(); err != nil {
		return nil, err
	}

	return proxy, nil
}

func NewFromFile(config string) (*DNSProxy, error) {
	return New(&ConfigYAMLFileLoader{path: config})
}
