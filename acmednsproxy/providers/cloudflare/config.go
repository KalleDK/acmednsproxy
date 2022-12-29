package cloudflare

import (
	"errors"
	"io"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v2"
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

type yamlConfig struct {
	Type   providers.Type
	Config `yaml:",inline"`
}

func LoadFromDecoder(dec auth.Decoder) (p *DNSProvider, err error) {
	var config yamlConfig
	if err = dec.Decode(&config); err != nil {
		return
	}

	if config.Type != Cloudflare {
		return nil, errors.New("invalid cloudflare " + string(config.Type))
	}
	return New(config.Config)
}

func LoadFromStream(r io.Reader) (d *DNSProvider, err error) {
	return LoadFromDecoder(yaml.NewDecoder(r))
}

func LoadFromFile(path string) (p *DNSProvider, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}

func wrappedLoad(dec providers.Decoder) (providers.DNSProvider, error) {
	return LoadFromDecoder(dec)
}

func init() {
	Cloudflare.Register(wrappedLoad)
}
