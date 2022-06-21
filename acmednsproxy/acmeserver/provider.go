package acmeserver

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

type DNSProviderLoader interface {
	Load(d ConfigDecoder) (p providers.DNSProvider, err error)
}

type DNSProvider string

var providerMap = map[DNSProvider]DNSProviderLoader{}

func RegisterDNSProvider(name DNSProvider, l DNSProviderLoader) {
	providerMap[name] = l
}

func (pl DNSProvider) Load(d ConfigDecoder) (p providers.DNSProvider, err error) {
	return providerMap[pl].Load(d)
}
