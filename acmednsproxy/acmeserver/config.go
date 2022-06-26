package acmeserver

import (
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

type ConfigDecoder interface {
	Decode(v interface{}) error
}

type ConfigFiles struct {
	DNSType  providers.DNSProviderName
	DNSPath  string
	AuthType Authenticator
	AuthPath string
}

func (c ConfigFiles) LoadAuth() (p auth.Authenticator, err error) {
	fp, err := os.Open(c.AuthPath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	dec := yaml.NewDecoder(fp)

	if p, err = c.AuthType.Load(dec); err != nil {
		return nil, err
	}

	return p, nil
}

func (c ConfigFiles) LoadProvider() (p providers.DNSProvider, err error) {
	fp, err := os.Open(c.DNSPath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	dec := yaml.NewDecoder(fp)

	if p, err = c.DNSType.Load(dec); err != nil {
		return nil, err
	}

	return p, nil
}
