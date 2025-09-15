package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

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
	domains map[string]struct{}
}

func (d *DNSProvider) CreateRecord(record providers.Record) error {
	token := record.Token()

	recordID, err := d.api.CreateDNSRecord(record.Fqdn, record.Value)
	if err != nil {
		return err
	}

	d.records.Add(token, recordID)

	log.Printf("cloudflare: new record for %s, ID %s", record.Fqdn, recordID)

	return nil
}

func (d *DNSProvider) RemoveRecord(record providers.Record) error {
	token := record.Token()

	recordID, err := d.records.Get(token)
	if err != nil {
		return fmt.Errorf("cloudflare: unknown record ID for '%s'", record)
	}

	if err := d.api.DeleteDNSRecord(recordID, record.Fqdn); err != nil {
		log.Printf("cloudflare: failed to delete TXT record: %s", err)
	}

	d.records.Delete(token)

	return nil
}

func (d *DNSProvider) CanHandle(domain string) bool {
	_, ok := d.domains[domain]
	return ok
}

func (d *DNSProvider) Close() error { return nil }

func (d *DNSProvider) Shutdown(ctx context.Context) error {
	return d.Close()
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
		Zones:      config.Zones,
		TTL:        ttl,
		HTTPClient: http_client,
	}

	api, err := newAPIClient(&api_config)
	if err != nil {
		return nil, err
	}

	domains := map[string]struct{}{}
	for _, zone := range config.Zones {
		domains[zone] = struct{}{}
	}

	return &DNSProvider{
		api: api,
		records: recordDB{
			values: map[string]string{},
			mutex:  sync.Mutex{},
		},
		domains: domains,
	}, nil
}

var _ providers.DNSProvider = (*DNSProvider)(nil)
