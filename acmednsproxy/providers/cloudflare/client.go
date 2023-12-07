package cloudflare

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

const minTTL = 120

type APIConfig struct {
	AuthToken  string
	ZoneID     string
	TTL        int
	HTTPClient *http.Client
}

type apiClient struct {
	clientEdit *cloudflare.API // needs Zone/DNS/Edit permissions
	zoneID     *cloudflare.ResourceContainer
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

	return &apiClient{
		clientEdit: dns,
		TTL:        config.TTL,
		zoneID:     cloudflare.ZoneIdentifier(config.ZoneID),
	}, nil
}

func (m *apiClient) CreateDNSRecord(fqdn, value string) (string, error) {
	dnsRecord := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    dns01.UnFqdn(fqdn),
		Content: value,
		TTL:     m.TTL,
	}

	response, err := m.clientEdit.CreateDNSRecord(context.Background(), m.zoneID, dnsRecord)
	if err != nil {
		return "", fmt.Errorf("cloudflare: failed to create TXT record: %w", err)
	}

	return response.ID, nil
}

func (m *apiClient) DeleteDNSRecord(recordID string) error {
	return m.clientEdit.DeleteDNSRecord(context.Background(), m.zoneID, recordID)
}
