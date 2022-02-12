package cloudflare

import (
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
)

type CFConfig struct {
	ZONE_API_TOKEN      string
	DNS_API_TOKEN       string
	POLLING_INTERVAL    *int
	PROPAGATION_TIMEOUT *int
	TTL                 *int
	HTTP_TIMEOUT        *int
}

func Load(dec providers.ConfigDecoder) (challenge.Provider, error) {
	var config CFConfig
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	conf := cloudflare.NewDefaultConfig()
	conf.AuthToken = config.DNS_API_TOKEN
	conf.ZoneToken = config.ZONE_API_TOKEN
	if config.TTL != nil {
		conf.TTL = *config.TTL
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

	p, err := cloudflare.NewDNSProviderConfig(conf)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func init() {
	providers.AddLoader("cloudflare", providers.LoaderFunc(Load))
}
