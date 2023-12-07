package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

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

	if err := d.api.DeleteDNSRecord(recordID); err != nil {
		log.Printf("cloudflare: failed to delete TXT record: %s", err)
	}

	d.records.Delete(token)

	return nil
}

func (d *DNSProvider) Close() error { return nil }

func (d *DNSProvider) Shutdown(ctx context.Context) error {
	return d.Close()
}
