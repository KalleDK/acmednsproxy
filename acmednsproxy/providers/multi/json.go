package multi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"github.com/go-acme/lego/v4/challenge"
)

type JsonConfigDecoder struct {
	Reader io.Reader
}

type entry struct {
	Name   string
	Config json.RawMessage
}

func (j JsonConfigDecoder) Decode(v interface{}) (err error) {

	vp, ok := v.(*map[string]challenge.Provider)
	if !ok {
		return errors.New("invalid type")
	}

	var entries map[string]entry

	if err = json.NewDecoder(j.Reader).Decode(&entries); err != nil {
		return
	}

	pm := map[string]challenge.Provider{}

	for domain, conf := range entries {
		pdec := json.NewDecoder(bytes.NewReader([]byte(conf.Config)))

		provider, err := providers.Load(conf.Name, pdec)
		if err != nil {
			return err
		}

		pm[domain] = provider
	}

	*vp = pm

	return nil
}
