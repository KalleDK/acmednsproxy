package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers/multi"
	"gopkg.in/yaml.v3"
)

const conf_str = `---
- domain: example.com
  type: cloudflare
  config:
    zone_id: Zpe7L9jRCkag6PuNC6b5a17gMtf156Ozq7tsRfhk
    dns_api_token: 7dZbonk5j5nDAq51Wd7nAuDF4rJl3FTQ_xYIrrKx
`

func main() {
	buf := bytes.NewBufferString(conf_str)

	dec := yaml.NewDecoder(buf)

	dnstype := providers.DNSProviderName("multi")

	p, err := dnstype.Load(dec)
	if err != nil {
		log.Fatal(err)
	}

	y := p.(*multi.MultiProvider)
	for k, v := range y.Providers {
		fmt.Printf("%s: %+v\n", k, v)
	}

}
