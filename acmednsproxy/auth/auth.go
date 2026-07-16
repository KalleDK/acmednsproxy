// Package auth defines the Authenticator interface and the type registry used
// to load authenticator implementations by name from YAML configuration files.
package auth

import (
	"context"
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Credentials holds the username and password supplied by a client in an
// HTTP Basic-Auth request.
type Credentials struct {
	Username string
	Password string
}

// Authenticator is implemented by every auth backend (e.g. simpleauth, noauth).
// It verifies that a set of credentials is authorised to manage DNS records for
// a given domain and exposes lifecycle hooks for orderly shutdown.
type Authenticator interface {
	io.Closer
	Shutdown(ctx context.Context) error
	// VerifyPermissions returns nil when cred is allowed to act on domain,
	// or a non-nil error (wrapping ErrUnauthorized, ErrUnknownUser, or
	// ErrUnknownDomain) otherwise.
	VerifyPermissions(cred Credentials, domain string) (err error)
}

// Decoder abstracts a YAML decoder so that authenticator loaders can be
// tested without touching the file system.
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder abstracts a YAML encoder used when serialising authenticator state.
type Encoder interface {
	Encode(v interface{}) error
}

// Loader is the function signature that every authenticator backend must
// register.  It receives the raw YAML node for the authenticator block and
// returns a fully initialised Authenticator.
type Loader func(dec Decoder) (Authenticator, error)

// Type is the string tag written in the YAML config (e.g. "simpleauth",
// "noauth") that identifies which backend to instantiate.
type Type string

// authenticatorMap is the global registry that maps Type strings to their
// corresponding Loader functions.  Backends register themselves via init().
var authenticatorMap = map[Type]Loader{}

// Register adds loader to the global registry under the given Type.
// It is called from each backend's init() function.
func (t Type) Register(loader Loader) {
	authenticatorMap[t] = loader
}

// LoadFromDecoder instantiates the authenticator backend identified by t and
// feeds it the YAML decoder dec.
func (u Type) LoadFromDecoder(dec Decoder) (a Authenticator, err error) {
	loader, ok := authenticatorMap[u]
	if !ok {
		return nil, errors.New("invalid authenticator " + string(u))
	}
	return loader(dec)
}

// LoadFromStream reads a YAML-encoded authenticator config from r and
// delegates to the registered backend identified by the "type" key.
func (u Type) LoadFromStream(r io.Reader) (c Authenticator, err error) {
	return u.LoadFromDecoder(yaml.NewDecoder(r))
}

// LoadFromFile reads the YAML file at path and delegates to the registered
// backend identified by the "type" key.
func (u Type) LoadFromFile(path string) (c Authenticator, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return u.LoadFromStream(r)
}

// yamlConfig is used to peek at the "type" field before dispatching to the
// correct backend loader.
type yamlConfig struct {
	Type Type `yaml:"type"`
}

// LoadFromDecoder reads a YAML node from dec, inspects the "type" field, and
// delegates to the matching registered Loader.
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

// LoadFromStream reads a complete authenticator config (including its "type"
// discriminator) from r and returns the corresponding Authenticator.
func LoadFromStream(r io.Reader) (c Authenticator, err error) {
	return LoadFromDecoder(yaml.NewDecoder(r))
}

// LoadFromFile reads the YAML file at path, inspects the "type" field, and
// returns the corresponding Authenticator.
func LoadFromFile(path string) (c Authenticator, err error) {
	r, err := os.Open(path)
	if err != nil {
		return
	}
	return LoadFromStream(r)
}
