package acmeserver

import (
	"log"
	"net/http"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
)

type Settings struct {
	Addr string
}

type TLSSettings struct {
	Addr     string
	CertFile string
	KeyFile  string
}

func Serve(server *acmeservice.DNSProxy, settings Settings) {
	handler, err := NewHandler(server)
	if err != nil {
		log.Panic(err)
	}

	if err := http.ListenAndServe(settings.Addr, handler); err != nil {
		log.Panic(err)
	}
}

func ServeTLS(server *acmeservice.DNSProxy, settings TLSSettings) {
	handler, err := NewHandler(server)
	if err != nil {
		log.Panic(err)
	}

	if err := http.ListenAndServeTLS(settings.Addr, settings.CertFile, settings.KeyFile, handler); err != nil {
		log.Panic(err)
	}
}
