package simpleauth

import (
	"errors"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"golang.org/x/exp/slices"
)

const NoAuth = auth.Type("noauth")

type Config struct {
	Domains []string
}

type Authenticator struct {
	domains []string
}

func (a *Authenticator) VerifyPermissions(user string, password string, domain string) (err error) {
	if slices.Contains(a.domains, domain) {
		return nil
	}

	return errors.New("not allowed")
}

func (a *Authenticator) Close() (err error) { return nil }

func (a *Authenticator) Load(config Config) {
	a.domains = config.Domains
}

func FromConfig(config Config) (*Authenticator, error) {
	var auth Authenticator
	auth.Load(config)
	return &auth, nil
}

func Load(unmarshal auth.YAMLUnmarshaler, config_dir string) (auth.Authenticator, error) {
	var conf Config
	if err := unmarshal(&conf); err != nil {
		return nil, err
	}

	return FromConfig(conf)
}

func init() {
	NoAuth.Register(Load)
}
