package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeserver"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

const (
	minTTL                    = 120
	defaultTTL                = minTTL
	defaultPropagationTimeout = 2 * time.Minute
	defaultPollingInterval    = 2 * time.Second
	defaultHTTPTimeout        = 30 * time.Second
)

type CFConfig struct {
	ZONE_ID             string
	DNS_API_TOKEN       string
	POLLING_INTERVAL    *int
	PROPAGATION_TIMEOUT *int
	TTL                 *int
	HTTP_TIMEOUT        *int
}

type loader struct{}

func updateIntIfExists(o *int, s *int) {
	if s != nil {
		*o = *s
	}
}

func updateSecondIfExists(d *time.Duration, s *int) {
	if s != nil {
		*d = time.Second * time.Duration(*s)
	}
}

func (l loader) Load(dec acmeserver.ConfigDecoder) (providers.DNSProvider, error) {
	var config CFConfig
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	conf := NewDefaultConfig()
	conf.API.AuthToken = config.DNS_API_TOKEN
	conf.API.ZoneID = config.ZONE_ID

	updateIntIfExists(&conf.Provider.TTL, config.TTL)
	updateSecondIfExists(&conf.Provider.PollingInterval, config.POLLING_INTERVAL)
	updateSecondIfExists(&conf.Provider.PropagationTimeout, config.PROPAGATION_TIMEOUT)
	updateSecondIfExists(&conf.API.HTTPClient.Timeout, config.HTTP_TIMEOUT)

	p, err := NewDNSProviderConfig(conf)
	if err != nil {
		return nil, err
	}

	return p, nil
}

var defaultLoader = loader{}
var providerName = acmeserver.DNSProvider("cloudflare")

func init() {
	acmeserver.RegisterDNSProvider(providerName, defaultLoader)
}

type ProviderConfig struct {
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	API      APIConfig
	Provider ProviderConfig
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Provider: ProviderConfig{
			TTL:                defaultTTL,
			PropagationTimeout: defaultPropagationTimeout,
			PollingInterval:    defaultPollingInterval,
		},
		API: APIConfig{
			HTTPClient: &http.Client{
				Timeout: defaultHTTPTimeout,
			},
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	api    *apiClient
	config ProviderConfig

	records recordDB
}

// NewDNSProviderConfig return a DNSProvider instance configured for Cloudflare.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cloudflare: the configuration of the DNS provider is nil")
	}

	if config.Provider.TTL < minTTL {
		return nil, fmt.Errorf("cloudflare: invalid TTL, TTL (%d) must be greater than %d", config.Provider.TTL, minTTL)
	}

	client, err := newAPIClient(&config.API)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %w", err)
	}

	return &DNSProvider{
		api:    client,
		config: config.Provider,
		records: recordDB{
			values: map[string]string{},
		},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) CreateRecord(fqdn, value string) error {
	token := fqdn + value

	dnsRecord := cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}

	response, err := d.api.CreateDNSRecord(context.Background(), dnsRecord)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to create TXT record: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("cloudflare: failed to create TXT record: %+v %+v", response.Errors, response.Messages)
	}

	d.records.Add(token, response.Result.ID)

	log.Printf("cloudflare: new record for %s, ID %s", fqdn, response.Result.ID)

	return nil
}

func (d *DNSProvider) RemoveRecord(fqdn, value string) error {
	token := fqdn + value

	recordID, err := d.records.Get(token)
	if err != nil {
		return fmt.Errorf("cloudflare: unknown record ID for '%s'", fqdn)
	}

	if err := d.api.DeleteDNSRecord(context.Background(), recordID); err != nil {
		log.Printf("cloudflare: failed to delete TXT record: %s", err)
	}

	d.records.Delete(token)

	return nil
}

type recordDB struct {
	values map[string]string
	mutex  sync.Mutex
}

func (r *recordDB) Get(name string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	v, ok := r.values[name]
	if !ok {
		return "", errors.New("missing token")
	}

	return v, nil
}

func (r *recordDB) Add(name, value string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.values[name] = value
}

func (r *recordDB) Delete(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.values, name)
}
