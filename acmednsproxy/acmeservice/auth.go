package acmeservice

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
)

type Auth struct {
	Username string
	Password string
}

type AuthConfig struct {
	Type   auth.Type
	Config RawYAML
}

func (ac AuthConfig) Load(config_dir string) (auth.Authenticator, error) {
	return ac.Type.Load(ac.Config.unmarshal, config_dir)
}
