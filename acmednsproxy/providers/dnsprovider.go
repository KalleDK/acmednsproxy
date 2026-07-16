// Package providers defines the DNSProvider interface, the global provider
// registry, and the DNSProviders multiplexer that routes DNS challenge records
// to the correct backend based on domain name.
package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"gopkg.in/yaml.v3"
)

// Record represents a single DNS TXT challenge record to be created or
// removed.  Fqdn is the fully-qualified domain name (with trailing dot) and
// Value is the TXT record content.
type Record struct {
	Fqdn  string
	Value string
}

// Token returns a stable string key that uniquely identifies a Record.
// It is used internally to correlate a CreateRecord call with the subsequent
// RemoveRecord call.
func (r Record) Token() string {
	return fmt.Sprintf("%s=%s", r.Fqdn, r.Value)
}

// DNSProvider is implemented by every DNS backend (e.g. cloudflare, httpreq).
// Backends must be safe for concurrent use.
type DNSProvider interface {
	io.Closer
	Shutdown(ctx context.Context) error
	// CreateRecord publishes a TXT record for the ACME DNS-01 challenge.
	CreateRecord(record Record) error
	// RemoveRecord deletes a previously published TXT record.
	RemoveRecord(record Record) error
	// CanHandle reports whether this provider is responsible for domain.
	// domain is an unqualified domain name (no trailing dot).
	CanHandle(domain string) bool
}

// DNSProviderLoader is the function signature that every backend must register.
// It receives the raw YAML node for the provider block and returns a fully
// initialised DNSProvider.
type DNSProviderLoader func(dec *yaml.Node) (DNSProvider, error)

// Type is the string tag written in the YAML config (e.g. "cloudflare",
// "httpreq") that identifies which backend to instantiate.
type Type string

// providerMap is the global registry that maps Type strings to their
// corresponding DNSProviderLoader functions.  Backends register via init().
var providerMap = map[Type]DNSProviderLoader{}

// Register adds loader to the global registry under t.  Panics if t is
// already registered to catch accidental duplicate imports.
func Register(t Type, loader DNSProviderLoader) {
	if _, exists := providerMap[t]; exists {
		panic("provider " + string(t) + " already registered")
	}
	providerMap[t] = loader
}

// load looks up the registered loader for t and uses it to instantiate a
// DNSProvider from the YAML node dec.
func (t Type) load(dec *yaml.Node) (p DNSProvider, err error) {
	loader, ok := providerMap[t]
	if !ok {
		return nil, errors.New("invalid provider " + string(t))
	}
	return loader(dec)
}

// yamlConfig is used to peek at the "type" field before dispatching to the
// correct backend loader.
type yamlConfig struct {
	Type Type `yaml:"type"`
}

// loadFromDecoder reads a YAML list of provider configs from dec, instantiates
// each backend, and returns a DNSProviders multiplexer.
func loadFromDecoder(dec *yaml.Decoder) (*DNSProviders, error) {
	var rawConfigs []yaml.Node
	provider := DNSProviders{}

	if err := dec.Decode(&rawConfigs); err != nil {
		return nil, err
	}

	for _, node := range rawConfigs {
		var config yamlConfig
		if err := node.Decode(&config); err != nil {
			return nil, err
		}
		subprovider, err := config.Type.load(&node)
		if err != nil {
			return nil, err
		}
		provider.providers = append(provider.providers, subprovider)
	}

	return &provider, nil
}

// LoadFromStream reads a YAML-encoded list of provider configs from r and
// returns a DNSProviders multiplexer.
func LoadFromStream(r io.Reader) (p *DNSProviders, err error) {
	return loadFromDecoder(yaml.NewDecoder(r))
}

// LoadFromFile reads the YAML file at path and returns a DNSProviders
// multiplexer.
func LoadFromFile(path string) (p *DNSProviders, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}

// DNSProviders is a multiplexer that dispatches DNS challenge operations to
// the appropriate backend based on the domain name of the challenge record.
type DNSProviders struct {
	providers []DNSProvider
}

// getProvider returns the first registered DNSProvider whose CanHandle method
// returns true for domain or any of its parent domains.  The search starts
// with the most-specific label and progressively strips leading labels until
// a match is found or all labels are exhausted.
func (mp *DNSProviders) getProvider(domain string) (p DNSProvider, err error) {
	domainParts := strings.Split(domain, ".")
	for len(domainParts) > 0 {
		domainStub := strings.Join(domainParts, ".")
		for _, sp := range mp.providers {
			if sp.CanHandle(domainStub) {
				return sp, nil
			}
		}
		domainParts = domainParts[1:]
	}
	return nil, errors.New("no matching provider")
}

// RemoveRecord strips the trailing dot from record.Fqdn, resolves the
// responsible backend, and delegates the removal.
func (mp *DNSProviders) RemoveRecord(record Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	sp, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = sp.RemoveRecord(record); err != nil {
		return err
	}

	return nil
}

// CreateRecord strips the trailing dot from record.Fqdn, resolves the
// responsible backend, and delegates the creation.
func (mp *DNSProviders) CreateRecord(record Record) error {
	domain := dns01.UnFqdn(record.Fqdn)
	sp, err := mp.getProvider(domain)
	if err != nil {
		return err
	}

	if err = sp.CreateRecord(record); err != nil {
		return err
	}

	return nil
}

// Close shuts down all registered backends.  All errors are collected and
// joined so that a failure in one backend does not prevent the others from
// being closed.
func (mp *DNSProviders) Close() error {
	var errs []error
	for _, sp := range mp.providers {
		if err := sp.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Shutdown calls Close on all registered backends.
func (mp *DNSProviders) Shutdown(ctx context.Context) error {
	return mp.Close()
}
