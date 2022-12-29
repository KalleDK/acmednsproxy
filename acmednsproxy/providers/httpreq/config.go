package httpreq

import (
	"errors"
	"io"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v2"
)

const HTTPREQ = providers.Type("httpreq")

type Config struct {
	Endpoint    string
	Username    string
	Password    string
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

	if config.Type != HTTPREQ {
		return nil, errors.New("invalid httpreq provider " + string(config.Type))
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
	HTTPREQ.Register(wrappedLoad)
}
