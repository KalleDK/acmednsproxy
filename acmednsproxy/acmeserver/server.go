package acmeserver

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/httphandlers"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/cloudflare"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/httpreq"
)

func loadAuthFile(authFile string) (a auth.SimpleUserAuthenticator, err error) {
	data, err := os.ReadFile(authFile)
	if err != nil {
		return
	}

	if err = a.Load(bytes.NewReader(data)); err != nil {
		return
	}

	return a, nil
}

func loadproviderFile(providerFile string) (p providers.Providers, err error) {
	data, err := os.ReadFile(providerFile)
	if err != nil {
		return
	}

	p, err = providers.LoadProviders(bytes.NewReader(data))
	if err != nil {
		return
	}

	return p, nil
}

type Server struct {
	AuthFile     string
	ProviderFile string
	CertFile     string
	KeyFile      string
	Auth         auth.SimpleUserAuthenticator
	Providers    providers.Providers
}

func (s *Server) Serve() {
	handler, err := httphandlers.NewHandler(&s.Auth, &s.Providers)
	if err != nil {
		log.Panic(err)
	}

	if s.CertFile == "" {
		http.ListenAndServe(":8080", handler)
	} else {
		http.ListenAndServeTLS(":9090", s.CertFile, s.KeyFile, handler)
	}

}

func (s *Server) ReloadConfig() (err error) {
	data, err := os.ReadFile(s.AuthFile)
	if err != nil {
		return
	}

	if err = s.Auth.Load(bytes.NewReader(data)); err != nil {
		return
	}

	return nil
}

func New(authfile string, providerfile string, certFile string, keyFile string) Server {

	p, err := loadproviderFile(providerfile)
	if err != nil {
		log.Panic(err)
	}

	a, err := loadAuthFile(authfile)
	if err != nil {
		log.Panic(err)
	}

	return Server{
		AuthFile:     authfile,
		ProviderFile: providerfile,
		CertFile:     certFile,
		KeyFile:      keyFile,
		Auth:         a,
		Providers:    p,
	}
}
