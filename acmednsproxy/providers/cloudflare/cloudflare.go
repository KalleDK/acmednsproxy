// Package cloudflare implements the providers.DNSProvider interface using the
// Cloudflare API to manage DNS TXT records for ACME DNS-01 challenges.
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

// recordDB is a thread-safe in-memory map from Record.Token() strings to the
// Cloudflare record IDs returned by CreateDNSRecord.  The IDs are needed when
// deleting records later.
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

// DNSProvider implements providers.DNSProvider using the Cloudflare API.
// It stores the Cloudflare record IDs of created records so they can be
// removed later.
type DNSProvider struct {
	api     *apiClient
	records recordDB
	// domains is the set of unqualified domain names this provider handles,
	// derived from the zone keys in Config.Zones.
	domains map[string]struct{}
}

// CreateRecord publishes a TXT record via the Cloudflare API and caches the
// returned record ID so it can be removed by RemoveRecord.
func (d *DNSProvider) CreateRecord(record providers.Record) error {
	token := record.Token()

	recordID, err := d.api.CreateDNSRecord(record.Fqdn, record.Value)
	if err != nil {
		return err
	}

	d.records.Add(token, recordID)

	log.Printf("cloudflare: created record for %s, ID %s", record.Fqdn, recordID)

	return nil
}

// RemoveRecord deletes a previously created TXT record from Cloudflare.
// It looks up the record ID in the in-memory cache by the record's token.
func (d *DNSProvider) RemoveRecord(record providers.Record) error {
	token := record.Token()

	recordID, err := d.records.Get(token)
	if err != nil {
		return fmt.Errorf("cloudflare: unknown record ID for '%s'", record)
	}

	if err := d.api.DeleteDNSRecord(recordID, record.Fqdn); err != nil {
		return fmt.Errorf("cloudflare: failed to delete TXT record: %w", err)
	}

	d.records.Delete(token)

	return nil
}

// CanHandle reports whether this provider is configured to handle domain.
// domain is compared against the unqualified zone names supplied in Config.Zones.
func (d *DNSProvider) CanHandle(domain string) bool {
	_, ok := d.domains[domain]
	return ok
}

// Close is a no-op; the Cloudflare HTTP client has no persistent connections
// that need to be torn down.
func (d *DNSProvider) Close() error { return nil }

// Shutdown calls Close.
func (d *DNSProvider) Shutdown(ctx context.Context) error {
	return d.Close()
}

// New constructs a DNSProvider from config.  TTL defaults to minTTL if not
// set.  The optional HTTPTimeout is applied to the HTTP client used for
// Cloudflare API calls.
func New(config Config) (*DNSProvider, error) {
	ttl := minTTL
	if config.TTL != nil {
		ttl = *config.TTL
	}

	httpClient := &http.Client{}
	if config.HTTPTimeout != nil {
		httpClient.Timeout = time.Second * time.Duration(*config.HTTPTimeout)
	}

	apiConfig := APIConfig{
		AuthToken:  config.AuthToken,
		Zones:      config.Zones,
		TTL:        ttl,
		HTTPClient: httpClient,
	}

	api, err := newAPIClient(&apiConfig)
	if err != nil {
		return nil, err
	}

	// Build the set of domain names this provider can handle from the zone
	// keys (not the values, which are zone IDs).
	domains := make(map[string]struct{}, len(config.Zones))
	for domain := range config.Zones {
		domains[domain] = struct{}{}
	}

	return &DNSProvider{
		api: api,
		records: recordDB{
			values: map[string]string{},
		},
		domains: domains,
	}, nil
}

var _ providers.DNSProvider = (*DNSProvider)(nil)
