// Package httpreq implements providers.DNSProvider by forwarding DNS challenge
// operations to another acmednsproxy instance (or any compatible HTTP endpoint)
// via HTTP POST requests.  This allows a single public-facing proxy to delegate
// to backend proxies that have direct access to the DNS provider credentials.
package httpreq

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

// HTTPREQ is the Type constant used to identify this backend in YAML
// configuration files.
const HTTPREQ = providers.Type("httpreq")

// Config holds the YAML-decoded settings for an httpreq provider instance.
type Config struct {
	// Endpoint is the base URL of the remote acmednsproxy (e.g.
	// "https://ns01.example.com:9090").
	Endpoint string
	// Username is the HTTP Basic-Auth username sent to the remote proxy.
	Username string
	// Password is the HTTP Basic-Auth password sent to the remote proxy.
	Password string
	// HTTPTimeout is the request timeout in seconds.  If nil, no timeout
	// is applied.
	HTTPTimeout *int
}

func loadFromDecoder(dec *yaml.Node) (p *DNSProvider, err error) {
	var config Config
	if err = dec.Decode(&config); err != nil {
		return
	}

	return New(config)
}

func load(dec *yaml.Node) (providers.DNSProvider, error) {
	return loadFromDecoder(dec)
}

func init() {
	providers.Register(HTTPREQ, load)
}
