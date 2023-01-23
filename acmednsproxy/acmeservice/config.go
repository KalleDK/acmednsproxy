package acmeservice

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type RawYAML struct {
	unmarshal func(interface{}) error
}

func (r *RawYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.unmarshal = unmarshal
	return nil
}

type Config struct {
	Authenticator AuthConfig
	Provider      ProviderConfig
}

func (c *Config) Load(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	dec := yaml.NewDecoder(fp)
	return dec.Decode(c)
}

type ConfigLoader interface {
	Load() (config Config, conf_dir string, err error)
}

type ConfigYAMLFileLoader struct {
	path string
}

func (c *ConfigYAMLFileLoader) Load() (config Config, conf_dir string, err error) {
	conf_dir = path.Dir(c.path)

	fp, err := os.Open(c.path)
	if err != nil {
		return config, conf_dir, err
	}
	defer fp.Close()

	dec := yaml.NewDecoder(fp)

	if err := dec.Decode(&config); err != nil {
		return config, conf_dir, err
	}

	return config, conf_dir, nil
}

type ConfigYAMLBytesLoader struct {
	Config    []byte
	ConfigDir string
}

func (c *ConfigYAMLBytesLoader) Load() (config Config, conf_dir string, err error) {

	if c.ConfigDir != "" {
		conf_dir = c.ConfigDir
	} else {
		conf_dir = "."
	}

	if err := yaml.Unmarshal(c.Config, &config); err != nil {
		return config, conf_dir, err
	}

	return config, conf_dir, nil
}
