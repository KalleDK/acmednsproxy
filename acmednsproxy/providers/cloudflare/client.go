package cloudflare

import (
	"context"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
)

type APIConfig struct {
	AuthToken  string
	ZoneID     string
	HTTPClient *http.Client
}

type apiClient struct {
	clientEdit *cloudflare.API // needs Zone/DNS/Edit permissions
	zoneID     string
}

func newAPIClient(config *APIConfig) (*apiClient, error) {
	dns, err := cloudflare.NewWithAPIToken(config.AuthToken, cloudflare.HTTPClient(config.HTTPClient))
	if err != nil {
		return nil, err
	}

	return &apiClient{
		clientEdit: dns,
		zoneID:     config.ZoneID,
	}, nil
}

func (m *apiClient) CreateDNSRecord(ctx context.Context, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
	return m.clientEdit.CreateDNSRecord(ctx, m.zoneID, rr)
}

func (m *apiClient) DNSRecords(ctx context.Context, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error) {
	return m.clientEdit.DNSRecords(ctx, m.zoneID, rr)
}

func (m *apiClient) DeleteDNSRecord(ctx context.Context, recordID string) error {
	return m.clientEdit.DeleteDNSRecord(ctx, m.zoneID, recordID)
}
