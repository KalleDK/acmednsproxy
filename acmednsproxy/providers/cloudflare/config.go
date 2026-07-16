package cloudflare

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	"gopkg.in/yaml.v3"
)

// Cloudflare is the Type constant used to identify this backend in YAML
// configuration files.
const Cloudflare = providers.Type("cloudflare")

// Config holds the YAML-decoded settings for a single Cloudflare DNS provider
// instance.
type Config struct {
	// Zones maps unqualified domain names (e.g. "example.com") to their
	// Cloudflare zone IDs.
	Zones map[string]string
	// AuthToken is a Cloudflare API token with Zone/DNS/Edit permissions.
	AuthToken string
	// TTL is the record time-to-live in seconds.  Defaults to minTTL (120).
	TTL *int
	// HTTPTimeout is the request timeout in seconds for Cloudflare API calls.
	// If nil, no timeout is applied.
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
	providers.Register(Cloudflare, load)
}
