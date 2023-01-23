package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"gopkg.in/yaml.v3"
)

const conf_str = `---
type: multi
config:
  - type: cloudflare
    domain: example.com
    config:
      ttl: 150
      zoneid: Zpe7L9jRCkag6PuNC6b5a17gMtf156Ozq7tsRfhk
      authtoken: 7dZbonk5j5nDAq51Wd7nAuDF4rJl3FTQ_xYIrrKx
  - type: cloudflare
    domain: sub.example.com
    config:
      ttl: 150
      zoneid: Zpe7L9jRCkag6PuNC6b5a17gMtf156Ozq7tsRfhk
      authtoken: 7dZbonk5j5nDAq51Wd7nAuDF4rJl3FTQ_xYIrrKx
`

func main() {
	buf := bytes.NewBufferString(conf_str)

	dec := yaml.NewDecoder(buf)

	var config acmeservice.ProviderConfig

	if err := dec.Decode(&config); err != nil {
		log.Fatal(err)
	}

	p, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", p)
}
