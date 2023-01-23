package providers

import (
	"errors"
	"io"
)

type YAMLUnmarshaler func(interface{}) error

type DNSProvider interface {
	io.Closer
	CreateRecord(fqdn, value string) error
	RemoveRecord(fqdn, value string) error
}

type DNSProviderLoader func(unmarshal YAMLUnmarshaler, config_dir string) (DNSProvider, error)

type Type string

func (u Type) Load(unmarshal YAMLUnmarshaler, config_dir string) (p DNSProvider, err error) {
	loader, ok := providerMap[u]
	if !ok {
		return nil, errors.New("invalid provider " + string(u))
	}
	return loader(unmarshal, config_dir)
}

func (t Type) Register(loader DNSProviderLoader) {
	providerMap[t] = loader
}

var providerMap = map[Type]DNSProviderLoader{}
