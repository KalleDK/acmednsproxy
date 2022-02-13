package multi

import (
	"errors"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type MultiProvider struct {
	providers map[string]providers.ProviderSolved
}

func (mp *MultiProvider) getProvider(domain string) (p providers.ProviderSolved, err error) {
	domain_parts := strings.Split(domain, ".")
	for len(domain_parts) > 0 {
		domain_stub := strings.Join(domain_parts, ".")
		p, ok := mp.providers[domain_stub]
		if ok {
			return p, nil
		}
		domain_parts = domain_parts[1:]
	}
	return nil, errors.New("no matching provider")
}

func (mp *MultiProvider) Present(domain, token, keyAuth string) error {
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.Present(domain, token, keyAuth); err != nil {
		return err
	}

	return nil
}

func (mp *MultiProvider) CleanUp(domain, token, keyAuth string) error {
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.CleanUp(domain, token, keyAuth); err != nil {
		return err
	}

	return nil
}

func (mp *MultiProvider) RemoveRecord(fqdn, value string) error {
	domain := dns01.UnFqdn(fqdn)
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.RemoveRecord(fqdn, value); err != nil {
		return err
	}

	return nil
}

func (mp *MultiProvider) CreateRecord(fqdn, value string) error {
	domain := dns01.UnFqdn(fqdn)
	p, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = p.CreateRecord(fqdn, value); err != nil {
		return err
	}

	return nil
}

func Load(d providers.ConfigDecoder) (providers.ProviderSolved, error) {
	mp := &MultiProvider{}
	if err := d.Decode(&mp.providers); err != nil {
		return nil, err
	}

	return mp, nil
}

func init() {
	providers.AddLoader("multi", providers.LoaderFunc(Load))
}
