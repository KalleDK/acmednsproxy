package providers

import (
	"errors"
	"log"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

type ProviderSolved interface {
	CreateRecord(fqdn, value string) error
	RemoveRecord(fqdn, value string) error
}

func Present(p ProviderSolved, domain, token, keyAuth string) error {
	log.Printf("token %s", token)
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	if err := p.CreateRecord(fqdn, value); err != nil {
		return err
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func CleanUp(p ProviderSolved, domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if err := p.RemoveRecord(fqdn, value); err != nil {
		return err
	}

	return nil
}

type ConfigDecoder interface {
	Decode(v interface{}) error
}

type Loader interface {
	Load(configDecoder ConfigDecoder) (ProviderSolved, error)
}

type loaderFunc struct {
	load func(configDecoder ConfigDecoder) (ProviderSolved, error)
}

func (f loaderFunc) Load(configDecoder ConfigDecoder) (p ProviderSolved, err error) {
	p, err = f.load(configDecoder)
	if err != nil {
		return
	}
	return p, nil
}

func LoaderFunc(f func(configDecoder ConfigDecoder) (ProviderSolved, error)) Loader {
	return loaderFunc{
		load: f,
	}
}

var providerMap = map[string]Loader{}

func AddLoader(name string, p Loader) {
	providerMap[name] = p
}

func GetLoader(name string) (p Loader, err error) {
	p, ok := providerMap[name]
	if !ok {
		return nil, errors.New("invalid provider name")
	}

	return p, nil
}

func Load(name string, configDecoder ConfigDecoder) (p ProviderSolved, err error) {
	loader, err := GetLoader(name)
	if err != nil {
		return nil, err
	}

	p, err = loader.Load(configDecoder)
	if err != nil {
		return
	}

	return p, nil
}
