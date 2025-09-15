package httpreq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	Endpoint   *url.URL
	Username   string
	Password   string
	HTTPClient *http.Client
}

func (d *DNSProvider) RemoveRecord(record providers.Record) error {
	msg := &message{
		FQDN:  record.Fqdn,
		Value: record.Value,
	}

	err := d.doPost("/cleanup", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

func (d *DNSProvider) CreateRecord(record providers.Record) error {
	msg := &message{
		FQDN:  record.Fqdn,
		Value: record.Value,
	}

	err := d.doPost("/present", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %w", err)
	}
	return nil
}

func (d *DNSProvider) doPost(uri string, msg any) error {
	reqBody := &bytes.Buffer{}

	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return err
	}

	newURI := path.Join(d.Endpoint.EscapedPath(), uri)
	endpoint, err := d.Endpoint.Parse(newURI)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.Username) > 0 && len(d.Password) > 0 {
		req.SetBasicAuth(d.Username, d.Password)
	}

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %w", resp.StatusCode, err)
		}

		return fmt.Errorf("%d: request failed: %v", resp.StatusCode, string(body))
	}

	return nil
}

func (d *DNSProvider) CanHandle(domain string) bool {
	return true
}

func (d *DNSProvider) Close() error { return nil }

func (d *DNSProvider) Shutdown(ctx context.Context) error {
	return d.Close()
}

type message struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

func New(config Config) (p *DNSProvider, err error) {
	url, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if config.HTTPTimeout != nil {
		client.Timeout = time.Second * time.Duration(*config.HTTPTimeout)
	}

	return &DNSProvider{
		Endpoint:   url,
		Username:   config.Username,
		Password:   config.Password,
		HTTPClient: client,
	}, nil
}

var _ providers.DNSProvider = (*DNSProvider)(nil)
