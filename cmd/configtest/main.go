package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

//go:embed acme.yml
var conf_str []byte

func main() {

	config_loader := &acmeservice.ConfigYAMLBytesLoader{Config: conf_str, ConfigDir: "."}

	fmt.Printf("%+v\n", config_loader)

	p, err := acmeservice.New(config_loader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", p)

	if err := p.Reload(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(dns01.GetRecord("sub.example.com", "token"))

	fmt.Println(p.Authenticate(acmeservice.Auth{Username: "test", Password: "test"}, acmeservice.Record{FQDN: "_acme-challenge.sub.example.com.", Value: "PEaenWxYddN6Q_NT1PiOYfz4EsZu7jRXRlpAsNpBU-A"}))
	fmt.Println(p.Authenticate(acmeservice.Auth{Username: "dsa", Password: "dsa"}, acmeservice.Record{FQDN: "example.com", Value: ""}))

	fmt.Printf("%+v\n", p)
}
