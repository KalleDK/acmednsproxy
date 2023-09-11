package multi

import (
	"errors"
	"io"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

type YamlConfigDecoder struct {
	Reader io.Reader
}

type yamlentry struct {
	Domain string
	Type   string
	Config yaml.Node
}

func (j YamlConfigDecoder) Decode(v interface{}) (err error) {

	vp, ok := v.(*map[string]providers.DNSProvider)
	if !ok {
		return errors.New("invalid type")
	}

	var entries []yamlentry

	if err = yaml.NewDecoder(j.Reader).Decode(&entries); err != nil {
		return
	}

	pm := map[string]providers.DNSProvider{}

	for _, entry := range entries {
		provider, err := providers.Load(entry.Type, &entry.Config)
		if err != nil {
			return err
		}

		pm[entry.Domain] = provider
	}

	*vp = pm

	return nil
}
