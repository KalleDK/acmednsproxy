package multi

import (
	"errors"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"gopkg.in/yaml.v3"
)

type MultiProvider struct {
	Providers map[string]providers.DNSProvider
}

func (mp *MultiProvider) getProvider(domain string) (p providers.DNSProvider, err error) {
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

type yamlentry struct {
	Domain string
	Type   providers.DNSProviderName
	Config yaml.Node
}

type loader struct{}

func (l loader) Load(d providers.ConfigDecoder) (providers.DNSProvider, error) {

	var entries []yamlentry

	if err := d.Decode(&entries); err != nil {
		return nil, err
	}

	providers := map[string]providers.DNSProvider{}
	for _, entry := range entries {
		p, err := entry.Type.Load(&entry.Config)
		if err != nil {
			return nil, err
		}
		providers[entry.Domain] = p
	}

	return &MultiProvider{
		Providers: providers,
	}, nil
}

var defaultLoader = loader{}
var providerName = providers.DNSProviderName("multi")

func init() {
	providers.RegisterDNSProvider(providerName, defaultLoader)
}
