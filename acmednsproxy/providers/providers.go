package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/go-acme/lego/v4/challenge"
)

type Decoder interface {
	Decode(v interface{}) error
}

type ProviderLoader interface {
	Load(d Decoder) (challenge.Provider, error)
}

var providerMap = map[string]ProviderLoader{}

func AddProviderLoader(name string, p ProviderLoader) {
	providerMap[name] = p
}

func GetProviderLoader(name string) (p ProviderLoader, err error) {
	p, ok := providerMap[name]
	if !ok {
		return nil, errors.New("invalid provider name")
	}

	return p, nil
}

type Providers struct {
	providers map[string]challenge.Provider
}

func (p *Providers) GetProvider(name string) (challenge.Provider, error) {
	parts := strings.Split(name, ".")
	for len(parts) > 0 {
		sname := strings.Join(parts, ".")
		pp, ok := p.providers[sname]
		if !ok {
			parts = parts[1:]
			continue
		}
		return pp, nil
	}
	return nil, errors.New("missing provider")
}

type entry struct {
	Name   string
	Config json.RawMessage
}

func LoadProviders(r io.Reader) (p Providers, err error) {
	var entries map[string]entry

	dec := json.NewDecoder(r)

	if err = dec.Decode(&entries); err != nil {
		return
	}

	p.providers = make(map[string]challenge.Provider)

	for domain, conf := range entries {
		loader, err := GetProviderLoader(conf.Name)
		if err != nil {
			return p, err
		}
		pdec := json.NewDecoder(bytes.NewReader([]byte(conf.Config)))
		provider, err := loader.Load(pdec)
		if err != nil {
			return p, err
		}
		p.providers[domain] = provider
	}
	return p, nil
}

type ProviderBackend interface {
	GetProvider(name string) (challenge.Provider, error)
}
