package cloudflare

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
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

// #region RecordDB

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

// #endregion

type DNSProvider struct {
	api     *apiClient
	records recordDB
}

func (d *DNSProvider) CreateRecord(fqdn, value string) error {
	token := fqdn + value

	recordID, err := d.api.CreateDNSRecord(fqdn, value)
	if err != nil {
		return err
	}

	d.records.Add(token, recordID)

	log.Printf("cloudflare: new record for %s, ID %s", fqdn, recordID)

	return nil
}

func (d *DNSProvider) RemoveRecord(fqdn, value string) error {
	token := fqdn + value

	recordID, err := d.records.Get(token)
	if err != nil {
		return fmt.Errorf("cloudflare: unknown record ID for '%s'", fqdn)
	}

	if err := d.api.DeleteDNSRecord(recordID); err != nil {
		log.Printf("cloudflare: failed to delete TXT record: %s", err)
	}

	d.records.Delete(token)

	return nil
}

func (d *DNSProvider) Close() error { return nil }

func FromConfig(config Config) (*DNSProvider, error) {

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

func Load(unmarshal providers.YAMLUnmarshaler, config_dir string) (providers.DNSProvider, error) {
	var conf Config
	if err := unmarshal(&conf); err != nil {
		return nil, err
	}

	return FromConfig(conf)
}

func init() {
	Cloudflare.Register(Load)
}
