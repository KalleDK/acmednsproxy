package providers

import "errors"

type ConfigDecoder interface {
	Decode(v interface{}) error
}

type DNSProviderLoader interface {
	Load(d ConfigDecoder) (p DNSProvider, err error)
}

type DNSProviderName string

var providerMap = map[DNSProviderName]DNSProviderLoader{}

func RegisterDNSProvider(name DNSProviderName, l DNSProviderLoader) {
	providerMap[name] = l
}

func (pl DNSProviderName) Load(d ConfigDecoder) (p DNSProvider, err error) {
	loader, ok := providerMap[pl]
	if !ok {
		return nil, errors.New("no provider by that name")
	}
	return loader.Load(d)
}
