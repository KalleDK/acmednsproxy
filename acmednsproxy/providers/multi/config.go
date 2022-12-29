package multi

import (
	"errors"
	"io"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

const Multi = providers.Type("multi")

type SubConfig struct {
	Domains []string
	Config  yaml.Node
}

type Config struct {
	Providers []SubConfig
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

	if config.Type != Multi {
		return nil, errors.New("invalid multi " + string(config.Type))
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
	Multi.Register(wrappedLoad)
}
