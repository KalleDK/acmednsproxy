package cloudflare

import (
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v2"
)

const (
	Cloudflare = providers.Type("cloudflare")
)

type Config struct {
	ZoneID      string
	AuthToken   string
	TTL         *int
	HTTPTimeout *int
}

func New(config Config) (*DNSProvider, error) {

	ttl := minTTL
	if config.TTL != nil {
		ttl = *config.TTL
	}

	http_client := &http.Client{}
	if config.HTTPTimeout != nil {
		http_client.Timeout = time.Second * time.Duration(*config.HTTPTimeout)
	}

	api_config := APIConfig{
		AuthToken:  config.AuthToken,
		ZoneID:     config.ZoneID,
		TTL:        ttl,
		HTTPClient: http_client,
	}

	api, err := newAPIClient(&api_config)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		api: api,
		records: recordDB{
			values: map[string]string{},
			mutex:  sync.Mutex{},
		},
	}, nil
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
