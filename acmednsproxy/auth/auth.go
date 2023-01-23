package auth

import (
	"errors"
	"io"
)

type Authenticator interface {
	io.Closer
	VerifyPermissions(user string, password string, domain string) (err error)
}

type YAMLUnmarshaler func(interface{}) error

type AuthenticatorLoader func(unmarshal YAMLUnmarshaler, config_dir string) (Authenticator, error)

type Type string

func (u Type) Load(unmarshal YAMLUnmarshaler, config_dir string) (c Authenticator, err error) {
	loader, ok := authenticatorMap[u]
	if !ok {
		return nil, errors.New("invalid authenticator " + string(u))
	}
	return loader(unmarshal, config_dir)
}

func (t Type) Register(loader AuthenticatorLoader) {
	authenticatorMap[t] = loader
}

var authenticatorMap = map[Type]AuthenticatorLoader{}
