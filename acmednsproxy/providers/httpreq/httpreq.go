package httpreq

import (
	"net/http"
	"net/url"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

const HTTPREQ = providers.Type("httpreq")

type message struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

type Config struct {
	Endpoint    string
	Username    string
	Password    string
	HTTPTimeout *int
}

func FromConfig(config Config) (*DNSProvider, error) {
	url, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if config.HTTPTimeout != nil {
		client.Timeout = time.Second * time.Duration(*config.HTTPTimeout)
	}

	return &DNSProvider{
		Endpoint:   url,
		Username:   config.Username,
		Password:   config.Password,
		HTTPClient: client,
	}, nil
}

func Load(unmarshal providers.YAMLUnmarshaler, config_dir string) (providers.DNSProvider, error) {
	var conf Config
	if err := unmarshal(&conf); err != nil {
		return nil, err
	}

	return FromConfig(conf)
}

func init() {
	HTTPREQ.Register(Load)
}
