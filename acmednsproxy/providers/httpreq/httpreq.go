package httpreq

import (
	"net/url"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/httpreq"
)

type CFConfig struct {
	Endpoint            string
	Mode                *string
	Username            *string
	Password            *string
	POLLING_INTERVAL    *int
	PROPAGATION_TIMEOUT *int
	HTTP_TIMEOUT        *int
}

type HTTPReqLoader struct{}

func (c HTTPReqLoader) Load(d providers.Decoder) (challenge.Provider, error) {
	var config CFConfig
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	conf := httpreq.NewDefaultConfig()

	url, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}

	conf.Endpoint = url
	conf.Mode = "RAW"

	if config.Username != nil {
		conf.Username = *config.Username
	}

	if config.Password != nil {
		conf.Password = *config.Password
	}

	if config.POLLING_INTERVAL != nil {
		conf.PollingInterval = time.Second * time.Duration(*config.POLLING_INTERVAL)
	}

	if config.PROPAGATION_TIMEOUT != nil {
		conf.PropagationTimeout = time.Second * time.Duration(*config.PROPAGATION_TIMEOUT)
	}

	if config.HTTP_TIMEOUT != nil {
		conf.HTTPClient.Timeout = time.Second * time.Duration(*config.HTTP_TIMEOUT)
	}

	p, err := httpreq.NewDNSProviderConfig(conf)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func init() {
	providers.AddProviderLoader("httpreq", HTTPReqLoader{})
}
