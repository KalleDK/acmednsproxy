package noauth

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
)

type Authenticator struct {
	domains map[string]struct{}
}

func (a *Authenticator) AddPermission(domain string) (err error) {
	if _, ok := a.domains[domain]; ok {
		return nil
	}
	a.domains[domain] = struct{}{}
	return nil
}

func (a *Authenticator) RemovePermission(domain string) (err error) {
	if _, ok := a.domains[domain]; ok {
		delete(a.domains, domain)
		return nil
	}
	return errors.New("unknown domain")
}

func (a *Authenticator) VerifyPermissions(cred auth.Credentials, domain string) (err error) {
	if _, ok := a.domains[domain]; ok {
		return nil
	}

	return errors.New("not allowed")
}

func (a *Authenticator) Close() (err error) { return nil }

func (a *Authenticator) Shutdown(ctx context.Context) error {
	return a.Close()
}

func (a *Authenticator) Load(config Config) error {
	a.domains = make(map[string]struct{}, len(config.Domains))
	for _, domain := range config.Domains {
		a.domains[domain] = struct{}{}
	}
	return nil
}

func (a *Authenticator) Save() (Config, error) {
	config := Config{}
	config.Domains = make([]string, 0, len(a.domains))
	for domain := range a.domains {
		config.Domains = append(config.Domains, domain)
	}
	slices.SortFunc(config.Domains, cmp.Compare)
	return config, nil
}

func New(config Config) (a *Authenticator, err error) {
	a = &Authenticator{}
	a.Load(config)
	return a, nil
}
