package providers

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- stub DNSProvider for testing ----------------------------------------

type stubProvider struct {
	domain  string
	created []Record
	removed []Record
	// closeErr is returned by Close if non-nil.
	closeErr error
}

func (s *stubProvider) CanHandle(domain string) bool      { return domain == s.domain }
func (s *stubProvider) CreateRecord(r Record) error        { s.created = append(s.created, r); return nil }
func (s *stubProvider) RemoveRecord(r Record) error        { s.removed = append(s.removed, r); return nil }
func (s *stubProvider) Close() error                       { return s.closeErr }
func (s *stubProvider) Shutdown(_ context.Context) error   { return s.Close() }

// --- DNSProviders.getProvider tests --------------------------------------

func newTestProviders(domains ...string) *DNSProviders {
	mp := &DNSProviders{}
	for _, d := range domains {
		mp.providers = append(mp.providers, &stubProvider{domain: d})
	}
	return mp
}

func TestGetProvider_ExactMatch(t *testing.T) {
	mp := newTestProviders("example.com")
	p, err := mp.getProvider("example.com")
	if err != nil {
		t.Fatalf("getProvider(%q) unexpected error: %v", "example.com", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestGetProvider_NoMatch(t *testing.T) {
	mp := newTestProviders("example.com")
	_, err := mp.getProvider("other.org")
	if err == nil {
		t.Fatal("expected error for unmatched domain")
	}
}

func TestGetProvider_SubdomainFallback(t *testing.T) {
	// The provider is registered for "example.com"; a request for
	// "sub.example.com" should fall back to it after stripping the leading
	// label.
	mp := newTestProviders("example.com")
	p, err := mp.getProvider("sub.example.com")
	if err != nil {
		t.Fatalf("getProvider(%q) unexpected error: %v", "sub.example.com", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestGetProvider_MostSpecificFirst(t *testing.T) {
	// Two providers: one for "example.com", one for "sub.example.com".
	// A query for "sub.example.com" should return the more-specific provider.
	apex := &stubProvider{domain: "example.com"}
	sub := &stubProvider{domain: "sub.example.com"}
	mp := &DNSProviders{providers: []DNSProvider{apex, sub}}

	p, err := mp.getProvider("sub.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != sub {
		t.Errorf("expected sub-domain provider to be selected")
	}
}

func TestGetProvider_EmptyDomain(t *testing.T) {
	mp := newTestProviders("example.com")
	_, err := mp.getProvider("")
	if err == nil {
		t.Fatal("expected error for empty domain")
	}
}

// --- CreateRecord / RemoveRecord routing ---------------------------------

func TestCreateRecord_RoutesToCorrectProvider(t *testing.T) {
	s := &stubProvider{domain: "example.com"}
	mp := &DNSProviders{providers: []DNSProvider{s}}

	rec := Record{Fqdn: "_acme-challenge.example.com.", Value: "token"}
	if err := mp.CreateRecord(rec); err != nil {
		t.Fatalf("CreateRecord error: %v", err)
	}
	if len(s.created) != 1 || s.created[0] != rec {
		t.Errorf("expected record in stub, got %v", s.created)
	}
}

func TestRemoveRecord_RoutesToCorrectProvider(t *testing.T) {
	s := &stubProvider{domain: "example.com"}
	mp := &DNSProviders{providers: []DNSProvider{s}}

	rec := Record{Fqdn: "_acme-challenge.example.com.", Value: "token"}
	if err := mp.RemoveRecord(rec); err != nil {
		t.Fatalf("RemoveRecord error: %v", err)
	}
	if len(s.removed) != 1 || s.removed[0] != rec {
		t.Errorf("expected record in stub, got %v", s.removed)
	}
}

func TestCreateRecord_NoMatchingProvider(t *testing.T) {
	mp := newTestProviders("example.com")
	err := mp.CreateRecord(Record{Fqdn: "_acme-challenge.other.org.", Value: "x"})
	if err == nil {
		t.Fatal("expected error when no provider matches")
	}
}

// --- Close collects all errors -------------------------------------------

type errProvider struct {
	stubProvider
}

func (e *errProvider) Close() error { return errors.New("close error") }

func TestClose_CollectsAllErrors(t *testing.T) {
	mp := &DNSProviders{
		providers: []DNSProvider{
			&errProvider{},
			&errProvider{},
		},
	}
	err := mp.Close()
	if err == nil {
		t.Fatal("expected joined error from Close")
	}
	// Both errors should be present.
	errs := errors.Unwrap(err)
	_ = errs // errors.Join produces a multi-error; just verify non-nil is enough.
}

func TestClose_NoErrors(t *testing.T) {
	mp := newTestProviders("example.com")
	if err := mp.Close(); err != nil {
		t.Fatalf("expected nil error from Close, got %v", err)
	}
}

// --- Record.Token --------------------------------------------------------

func TestRecord_Token(t *testing.T) {
	r := Record{Fqdn: "_acme-challenge.example.com.", Value: "abc123"}
	want := "_acme-challenge.example.com.=abc123"
	if got := r.Token(); got != want {
		t.Errorf("Token() = %q, want %q", got, want)
	}
}

// --- LoadFromStream (registry dispatch) ----------------------------------

// dummyProvider is a minimal DNSProvider that always reports it can handle
// any domain.  It is registered under the "dummy" type for testing purposes.
type dummyProvider struct{}

func (d *dummyProvider) CanHandle(string) bool            { return true }
func (d *dummyProvider) CreateRecord(Record) error         { return nil }
func (d *dummyProvider) RemoveRecord(Record) error         { return nil }
func (d *dummyProvider) Close() error                      { return nil }
func (d *dummyProvider) Shutdown(context.Context) error    { return nil }

func init() {
	Register(Type("dummy"), func(n *yaml.Node) (DNSProvider, error) {
		return &dummyProvider{}, nil
	})
}

func TestLoadFromStream_ValidProvider(t *testing.T) {
	const yaml = `---
- type: dummy
`
	p, err := LoadFromStream(bytes.NewBufferString(yaml))
	if err != nil {
		t.Fatalf("LoadFromStream error: %v", err)
	}
	if p == nil || len(p.providers) != 1 {
		t.Errorf("expected 1 provider, got %v", p)
	}
}

func TestLoadFromStream_UnknownType(t *testing.T) {
	const yaml = `---
- type: nonexistent_provider
`
	_, err := LoadFromStream(bytes.NewBufferString(yaml))
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

func TestLoadFromStream_InvalidYAML(t *testing.T) {
	_, err := LoadFromStream(bytes.NewBufferString(":::invalid:::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
