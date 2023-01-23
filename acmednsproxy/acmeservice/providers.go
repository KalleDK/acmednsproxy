package acmeservice

import "github.com/KalleDK/acmednsproxy/acmednsproxy/providers"

type Record struct {
	FQDN  string
	Value string
}

type ProviderConfig struct {
	Type   providers.Type
	Config RawYAML
}

func (pc ProviderConfig) Load(config_dir string) (providers.DNSProvider, error) {
	return pc.Type.Load(pc.Config.unmarshal, config_dir)
}
