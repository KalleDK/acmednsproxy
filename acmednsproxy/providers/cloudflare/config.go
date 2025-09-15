package cloudflare

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

const (
	Cloudflare = providers.Type("cloudflare")
)

type Config struct {
	Zones       map[string]string
	AuthToken   string
	TTL         *int
	HTTPTimeout *int
}

func loadFromDecoder(dec *yaml.Node) (p *DNSProvider, err error) {
	var config Config
	if err = dec.Decode(&config); err != nil {
		return
	}

	return New(config)
}

func load(dec *yaml.Node) (providers.DNSProvider, error) {
	return loadFromDecoder(dec)
}

func init() {
	providers.Register(Cloudflare, load)
}
