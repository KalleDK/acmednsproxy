package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"gopkg.in/yaml.v3"
)

// #region Record

type Record struct {
	Fqdn  string
	Value string
}

func (r Record) Token() string {
	return fmt.Sprintf("%s=%s", r.Fqdn, r.Value)
}

// #endregion

type DNSProvider interface {
	io.Closer
	Shutdown(ctx context.Context) error
	CreateRecord(record Record) error
	RemoveRecord(record Record) error
	CanHandle(domain string) bool
}

type DNSProviderLoader func(dec *yaml.Node) (DNSProvider, error)

// #region Type

type Type string

var providerMap = map[Type]DNSProviderLoader{}

func Register(t Type, loader DNSProviderLoader) {
	if _, exists := providerMap[t]; exists {
		panic("provider " + string(t) + " already registered")
	}
	providerMap[t] = loader
}

func (t Type) load(dec *yaml.Node) (p DNSProvider, err error) {
	loader, ok := providerMap[t]
	if !ok {
		return nil, errors.New("invalid provider " + string(t))
	}
	return loader(dec)
}

// #endregion

// #region Config

type yamlConfig struct {
	Type Type `yaml:"type"`
}

func loadFromDecoder(dec *yaml.Decoder) (*DNSProviders, error) {
	var raw_configs []yaml.Node
	provider := DNSProviders{}

	if err := dec.Decode(&raw_configs); err != nil {
		return nil, err
	}

	for _, node := range raw_configs {
		var config yamlConfig
		if err := node.Decode(&config); err != nil {
			return nil, err
		}
		subprovider, err := config.Type.load(&node)
		if err != nil {
			return nil, err
		}
		provider.providers = append(provider.providers, subprovider)
	}

	return &provider, nil
}

func LoadFromStream(r io.Reader) (p *DNSProviders, err error) {
	return loadFromDecoder(yaml.NewDecoder(r))
}

func LoadFromFile(path string) (p *DNSProviders, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}

// #endregion

type DNSProviders struct {
	providers []DNSProvider
}

func (mp *DNSProviders) getProvider(domain string) (p DNSProvider, err error) {
	domain_parts := strings.Split(domain, ".")
	for len(domain_parts) > 0 {
		domain_stub := strings.Join(domain_parts, ".")
		for _, sp := range mp.providers {
			if sp.CanHandle(domain_stub) {
				return sp, nil
			}
		}
		domain_parts = domain_parts[1:]
	}
	return nil, errors.New("no matching provider")
}

func (mp *DNSProviders) RemoveRecord(record Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	sp, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = sp.RemoveRecord(record); err != nil {
		return err
	}

	return nil
}

func (mp *DNSProviders) CreateRecord(record Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	sp, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = sp.CreateRecord(record); err != nil {
		return err
	}

	return nil
}

func (mp *DNSProviders) Close() error {
	var err error = nil
	for _, sp := range mp.providers {
		if err_s := sp.Close(); err_s != nil {
			err = err_s
		}
	}
	return err
}

func (mp *DNSProviders) Shutdown(ctx context.Context) error {
	return mp.Close()
}
