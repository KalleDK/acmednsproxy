package noauth

import (
	"errors"
	"io"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"gopkg.in/yaml.v3"
)

const NoAuth = auth.Type("noauth")

type Config struct {
	Domains []string
}

type yamlConfig struct {
	Type   auth.Type
	Config `yaml:",inline"`
}

func LoadFromDecoder(dec auth.Decoder) (a *Authenticator, err error) {
	var config yamlConfig
	if err = dec.Decode(&config); err != nil {
		return
	}

	if config.Type != NoAuth {
		return nil, errors.New("invalid authenticator " + string(config.Type))
	}
	return New(config.Config)
}

func LoadFromStream(r io.Reader) (a *Authenticator, err error) {
	return LoadFromDecoder(yaml.NewDecoder(r))
}

func LoadFromFile(path string) (a *Authenticator, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}

func wrappedLoad(dec auth.Decoder) (auth.Authenticator, error) {
	return LoadFromDecoder(dec)
}

func init() {
	NoAuth.Register(wrappedLoad)
}
