package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// minTTL is the lowest TTL (in seconds) accepted for DNS TXT records.
// Cloudflare enforces a minimum of 120 seconds.
const minTTL = 120

// APIConfig holds the parameters needed to construct an apiClient.
type APIConfig struct {
	// AuthToken is a Cloudflare API token with Zone/DNS/Edit permissions.
	AuthToken string
	// Zones maps unqualified domain names to their Cloudflare zone IDs.
	Zones map[string]string
	// TTL is the time-to-live in seconds for created TXT records.
	TTL int
	// HTTPClient is the HTTP client used for all Cloudflare API calls.
	// If nil, the default http.Client is used.
	HTTPClient *http.Client
}

// apiClient wraps the Cloudflare SDK and provides convenience methods for
// creating and deleting DNS TXT records.
type apiClient struct {
	clientEdit *cloudflare.API // requires Zone/DNS/Edit permissions
	// zones is a sorted (most-specific first) slice of qualified zone names.
	zones   []string
	zoneIDs map[string]*cloudflare.ResourceContainer
	TTL     int
}

// newAPIClient validates config and constructs a ready-to-use apiClient.
// Zones are stored with a trailing dot and sorted most-specific-first so that
// GetZoneID can find the longest matching suffix.
func newAPIClient(config *APIConfig) (*apiClient, error) {
	if config.TTL < minTTL {
		return nil, fmt.Errorf("cloudflare: ttl too low: minimum is %d, got %d", minTTL, config.TTL)
	}

	dns, err := cloudflare.NewWithAPIToken(config.AuthToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	zoneIDs := map[string]*cloudflare.ResourceContainer{}
	for domain, zoneid := range config.Zones {
		if !strings.HasSuffix(domain, ".") {
			domain = domain + "."
		}
		zoneIDs[domain] = cloudflare.ZoneIdentifier(zoneid)
	}
	zones := make([]string, 0, len(zoneIDs))
	for domain := range zoneIDs {
		zones = append(zones, domain)
	}
	SortDomains(zones)

	return &apiClient{
		clientEdit: dns,
		TTL:        config.TTL,
		zoneIDs:    zoneIDs,
		zones:      zones,
	}, nil
}

// GetZoneID returns the Cloudflare ResourceContainer for the zone that is
// responsible for domain.  Zones are searched most-specific-first (longest
// suffix match wins).
func (m *apiClient) GetZoneID(domain string) (*cloudflare.ResourceContainer, error) {
	for _, zone := range m.zones {
		if strings.HasSuffix(domain, zone) {
			return m.zoneIDs[zone], nil
		}
	}
	return nil, fmt.Errorf("cloudflare: no zone found for domain %s", domain)
}

// CreateDNSRecord publishes a TXT record for fqdn with the given value and
// returns the Cloudflare record ID that must be supplied to DeleteDNSRecord.
func (m *apiClient) CreateDNSRecord(fqdn, value string) (string, error) {
	zoneID, err := m.GetZoneID(fqdn)
	if err != nil {
		return "", err
	}

	dnsRecord := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     m.TTL,
	}

	response, err := m.clientEdit.CreateDNSRecord(context.Background(), zoneID, dnsRecord)
	if err != nil {
		return "", fmt.Errorf("cloudflare: failed to create TXT record: %w", err)
	}

	return response.ID, nil
}

// DeleteDNSRecord removes the TXT record identified by recordID from the zone
// responsible for fqdn.
func (m *apiClient) DeleteDNSRecord(recordID, fqdn string) error {
	zoneID, err := m.GetZoneID(fqdn)
	if err != nil {
		return err
	}
	return m.clientEdit.DeleteDNSRecord(context.Background(), zoneID, recordID)
}

// ReverseString returns s with its Unicode code points in reverse order.
// It is used by SortDomains to sort domain names by their reversed labels,
// achieving a longest-suffix-first ordering.
func ReverseString(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

// SortDomains sorts s in-place so that more-specific (longer) domain names
// appear before less-specific ones.  For example:
//
//	["example.com.", "sub.example.com."] → ["sub.example.com.", "example.com."]
//
// The algorithm reverses each domain string, sorts lexicographically (which
// groups labels by TLD first), then reverses back and reverses the whole slice
// to obtain most-specific-first order.
func SortDomains(s []string) {
	for i := range s {
		s[i] = ReverseString(s[i])
	}
	sort.Strings(s)
	for i := range s {
		s[i] = ReverseString(s[i])
	}
	for i := range s[:len(s)/2] {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
}
