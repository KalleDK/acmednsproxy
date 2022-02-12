package providers

import (
	"errors"

	"github.com/go-acme/lego/v4/challenge"
)

type ConfigDecoder interface {
	Decode(v interface{}) error
}

type Loader interface {
	Load(configDecoder ConfigDecoder) (challenge.Provider, error)
}

type loaderFunc struct {
	load func(configDecoder ConfigDecoder) (challenge.Provider, error)
}

func (f loaderFunc) Load(configDecoder ConfigDecoder) (p challenge.Provider, err error) {
	p, err = f.load(configDecoder)
	if err != nil {
		return
	}
	return p, nil
}

func LoaderFunc(f func(configDecoder ConfigDecoder) (challenge.Provider, error)) Loader {
	return loaderFunc{
		load: f,
	}
}

var providerMap = map[string]Loader{}

func AddLoader(name string, p Loader) {
	providerMap[name] = p
}

func GetLoader(name string) (p Loader, err error) {
	p, ok := providerMap[name]
	if !ok {
		return nil, errors.New("invalid provider name")
	}

	return p, nil
}

func Load(name string, configDecoder ConfigDecoder) (p challenge.Provider, err error) {
	loader, err := GetLoader(name)
	if err != nil {
		return nil, err
	}

	p, err = loader.Load(configDecoder)
	if err != nil {
		return
	}

	return p, nil
}
