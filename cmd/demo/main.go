package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

// An example showing how to unmarshal embedded
// structs from YAML.

type StructA struct {
	Name   string
	Type   string
	Config yaml.Node
}

var data = `
- name: kalle.dk
  type: cloudflare
  config:
    zone_api_token: "A"
    dns_api_token: B
- name: kalle.dev
  type: cloudflare
  config:
    zoneapitoken: "C"
    dns_api_token: D
`

type CFConfig struct {
	Zone_Api_Token string
	DNS_Api_Token  string
}

func main() {
	var b []StructA

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	cf := CFConfig{}
	var cf2 map[string]string

	err = b[0].Config.Decode(&cf)
	fmt.Println(err)
	fmt.Println(cf)

	err = b[0].Config.Decode(&cf2)
	fmt.Println(err)
	fmt.Println([]byte(b[0].Config.Value))

}
