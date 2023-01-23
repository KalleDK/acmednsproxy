package multi

import (
	"errors"
	"strings"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge/dns01"
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

func (mp *MultiProvider) Close() error {
	var err error = nil
	for _, p := range mp.Providers {
		if err_s := p.Close(); err_s != nil {
			err = err_s
		}
	}
	return err
}

var Multi = providers.Type("multi")

type RawYAML struct {
	unmarshal providers.YAMLUnmarshaler
}

func (r *RawYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.unmarshal = unmarshal
	return nil
}

type SubConfig struct {
	Domain string
	Type   providers.Type
	Config RawYAML
}

type Config []SubConfig

func FromConfig(config Config, config_dir string) (*MultiProvider, error) {
	p := &MultiProvider{
		Providers: map[string]providers.DNSProvider{},
	}
	for _, subconf := range config {
		sub_p, err := subconf.Type.Load(subconf.Config.unmarshal, config_dir)
		if err != nil {
			return nil, err
		}
		p.Providers[subconf.Domain] = sub_p
	}
	return p, nil
}

func Load(unmarshal providers.YAMLUnmarshaler, config_dir string) (providers.DNSProvider, error) {
	var conf Config
	if err := unmarshal(&conf); err != nil {
		return nil, err
	}

	return FromConfig(conf, config_dir)
}

func init() {
	Multi.Register(Load)
}
