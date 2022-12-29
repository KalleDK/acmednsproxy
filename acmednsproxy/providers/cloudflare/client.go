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

const minTTL = 120

type APIConfig struct {
	AuthToken  string
	Zones      map[string]string
	TTL        int
	HTTPClient *http.Client
}

type apiClient struct {
	clientEdit *cloudflare.API // needs Zone/DNS/Edit permissions
	zones      []string
	zoneIDs    map[string]*cloudflare.ResourceContainer
	TTL        int
}

func newAPIClient(config *APIConfig) (*apiClient, error) {

	if config.TTL < minTTL {
		return nil, fmt.Errorf("cloudflare: to low ttl min %d got %d", minTTL, config.TTL)
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
	zones := []string{}
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

func (m *apiClient) GetZoneID(domain string) (*cloudflare.ResourceContainer, error) {
	for _, zone := range m.zones {
		if strings.HasSuffix(domain, zone) {
			return m.zoneIDs[zone], nil
		}
	}
	return nil, fmt.Errorf("cloudflare: no zone found for domain %s", domain)
}

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

func (m *apiClient) DeleteDNSRecord(recordID, fqdn string) error {
	zoneID, err := m.GetZoneID(fqdn)
	if err != nil {
		return err
	}
	return m.clientEdit.DeleteDNSRecord(context.Background(), zoneID, recordID)
}

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
