package providers

import (
	"log"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

type DNSProvider interface {
	CreateRecord(fqdn, value string) error
	RemoveRecord(fqdn, value string) error
}

func Present(p DNSProvider, domain, token, keyAuth string) error {
	log.Printf("token %s", token)

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	if err := p.CreateRecord(fqdn, value); err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func CleanUp(p DNSProvider, domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if err := p.RemoveRecord(fqdn, value); err != nil {
		return err
	}

	return nil
}
