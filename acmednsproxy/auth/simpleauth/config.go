package simpleauth

import (
	"errors"
	"io"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"gopkg.in/yaml.v3"
)

const SimpleAuth = auth.Type("simpleauth")

type UserTable map[string]string

type PermissionTable map[string]UserTable

type Config struct {
	Permissions PermissionTable
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

	if config.Type != SimpleAuth {
		return nil, errors.New("invalid simpleauth auth " + string(config.Type))
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

func SaveToEncoder(enc auth.Encoder, a *Authenticator) (err error) {
	return enc.Encode(yamlConfig{
		Type: SimpleAuth,
		Config: Config{
			Permissions: a.Permissions,
		},
	})
}

func SaveToStream(w io.Writer, a *Authenticator) (err error) {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	err = SaveToEncoder(enc, a)
	return err
}

func SaveToFile(path string, a *Authenticator) (err error) {
	w, err := os.Create(path)
	if err != nil {
		return
	}
	defer w.Close()
	return SaveToStream(w, a)
}

func wrappedLoad(dec auth.Decoder) (auth.Authenticator, error) {
	return LoadFromDecoder(dec)
}

func init() {
	SimpleAuth.Register(wrappedLoad)
}
