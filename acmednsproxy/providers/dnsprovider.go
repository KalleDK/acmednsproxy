package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Record struct {
	Fqdn  string
	Value string
}

func (r Record) Token() string {
	return fmt.Sprintf("%s=%s", r.Fqdn, r.Value)
}

type DNSProvider interface {
	io.Closer
	Shutdown(ctx context.Context) error
	CreateRecord(record Record) error
	RemoveRecord(record Record) error
}

type Decoder interface {
	Decode(v interface{}) error
}

type Type string

type Loader func(dec Decoder) (DNSProvider, error)

var providerMap = map[Type]Loader{}

func (t Type) Register(loader Loader) {
	providerMap[t] = loader
}

func (u Type) LoadFromDecoder(dec Decoder) (p DNSProvider, err error) {
	loader, ok := providerMap[u]
	if !ok {
		return nil, errors.New("invalid provider " + string(u))
	}
	return loader(dec)
}

func (u Type) LoadFromStream(r io.Reader) (p DNSProvider, err error) {
	return u.LoadFromDecoder(yaml.NewDecoder(r))
}

func (u Type) LoadFromFile(path string) (p DNSProvider, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return u.LoadFromStream(r)
}

type yamlConfig struct {
	Type Type `yaml:"type"`
}

func LoadFromDecoder(dec Decoder) (p DNSProvider, err error) {
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

func LoadFromStream(r io.Reader) (p DNSProvider, err error) {
	return LoadFromDecoder(yaml.NewDecoder(r))
}

func LoadFromFile(path string) (p DNSProvider, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}
