package multi

import (
	"context"
	"errors"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type DNSProvider struct {
	Providers map[string]providers.DNSProvider
}

func (mp *DNSProvider) getProvider(domain string) (p providers.DNSProvider, err error) {
	domain_parts := strings.Split(domain, ".")
	for len(domain_parts) > 0 {
		domain_stub := strings.Join(domain_parts, ".")
		p, ok := mp.Providers[domain_stub]
		if ok {
			return p, nil
		}
		domain_parts = domain_parts[1:]
	}
	return nil, errors.New("no matching provider")
}

func (mp *DNSProvider) RemoveRecord(record providers.Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.RemoveRecord(record); err != nil {
		return err
	}

	return nil
}

func (mp *DNSProvider) CreateRecord(record providers.Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.CreateRecord(record); err != nil {
		return err
	}

	return nil
}

func (mp *DNSProvider) Close() error {
	var err error = nil
	for _, p := range mp.Providers {
		if err_s := p.Close(); err_s != nil {
			err = err_s
		}
	}
	return err
}

func (mp *DNSProvider) Shutdown(ctx context.Context) error {
	return mp.Close()
}

type typeWrapper struct {
	Type providers.Type
}

func loadSubProvider(subconf SubConfig) (providers.DNSProvider, error) {
	var t typeWrapper
	if err := subconf.Config.Decode(&t); err != nil {
		return nil, err
	}
	return t.Type.LoadFromDecoder(&subconf.Config)
}

func New(config Config) (*DNSProvider, error) {
	providers := map[string]providers.DNSProvider{}
	for _, subconf := range config.Providers {
		sp, err := loadSubProvider(subconf)
		if err != nil {
			return nil, err
		}
		if sp == nil {
			continue
		}
		for _, domain := range subconf.Domains {
			providers[domain] = sp
		}
	}

	return &DNSProvider{
		Providers: providers,
	}, nil
}
