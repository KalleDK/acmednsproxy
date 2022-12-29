package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

func main() {

	config := acmeservice.Config{
		Authenticator: "auth.yml",
		Provider:      "prov.yml",
	}

	p, err := acmeservice.New(config)
	if err != nil {
		log.Fatalf("new %v", err)
	}

	fmt.Printf("%+v\n", p)

	if err := p.Reload(); err != nil {
		log.Fatalf("reload %v", err)
	}

	fmt.Println(dns01.GetRecord("sub.example.com", "token"))

	fmt.Println(p.Authenticate(auth.Credentials{Username: "test", Password: "test"}, providers.Record{Fqdn: "_acme-challenge.sub.example.com.", Value: "PEaenWxYddN6Q_NT1PiOYfz4EsZu7jRXRlpAsNpBU-A"}))
	fmt.Println(p.Authenticate(auth.Credentials{Username: "dsa", Password: "dsa"}, providers.Record{Fqdn: "examdple.com", Value: ""}))

	fmt.Printf("%+v\n", p)
}
