package acmeserver

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/httphandlers"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers/multi"
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

func loadproviderFile(providerFile string) (p providers.ProviderSolved, err error) {
	data, err := os.ReadFile(providerFile)
	if err != nil {
		return
	}

	p, err = providers.Load("multi", multi.YamlConfigDecoder{Reader: bytes.NewReader(data)})
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
	Provider     providers.ProviderSolved
}

func (s *Server) Serve() {
	handler, err := httphandlers.NewHandler(&s.Auth, s.Provider)
	if err != nil {
		log.Panic(err)
	}

	if s.CertFile == "" {
		log.Print("http")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Panic(err)
		}
	} else {
		log.Print("https")
		if err := http.ListenAndServeTLS(":9090", s.CertFile, s.KeyFile, handler); err != nil {
			log.Panic(err)
		}
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
		Provider:     p,
	}
}
