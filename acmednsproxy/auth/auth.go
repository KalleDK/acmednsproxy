package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Credentials struct {
	Username string
	Password string
}

type Authenticator interface {
	io.Closer
	Shutdown(ctx context.Context) error
	VerifyPermissions(cred Credentials, domain string) (err error)
}

type Decoder interface {
	Decode(v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) error
}

type Loader func(dec Decoder) (Authenticator, error)

type Type string

var authenticatorMap = map[Type]Loader{}

func (t Type) Register(loader Loader) {
	authenticatorMap[t] = loader
}

func (u Type) LoadFromDecoder(dec Decoder) (a Authenticator, err error) {
	loader, ok := authenticatorMap[u]
	if !ok {
		fmt.Println("loader")
		return nil, errors.New("invalid authenticator " + string(u))
	}
	return loader(dec)
}

func (u Type) LoadFromStream(r io.Reader) (c Authenticator, err error) {
	return u.LoadFromDecoder(yaml.NewDecoder(r))
}

func (u Type) LoadFromFile(path string) (c Authenticator, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return u.LoadFromStream(r)
}

type yamlConfig struct {
	Type Type `yaml:"type"`
}

func LoadFromDecoder(dec Decoder) (a Authenticator, err error) {
	var node yaml.Node
	var config yamlConfig
	if err = dec.Decode(&node); err != nil {
		return
	}
	if err = node.Decode(&config); err != nil {
		return
	}
	return config.Type.LoadFromDecoder(&node)
}

func LoadFromStream(r io.Reader) (c Authenticator, err error) {
	return LoadFromDecoder(yaml.NewDecoder(r))
}

func LoadFromFile(path string) (c Authenticator, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}
