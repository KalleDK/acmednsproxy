package acmeserver

import (
	"log"
	"net/http"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/httphandlers"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
)

type Settings struct {
	Addr string
}

type TLSSettings struct {
	Addr     string
	CertFile string
	KeyFile  string
}

type Server struct {
	Config   ConfigFiles
	Auth     auth.Authenticator
	Provider providers.DNSProvider
}

func (s *Server) Serve(settings Settings) {
	handler, err := httphandlers.NewHandler(s.Auth, s.Provider)
	if err != nil {
		log.Panic(err)
	}

	if err := http.ListenAndServe(settings.Addr, handler); err != nil {
		log.Panic(err)
	}
}

func (s *Server) ServeTLS(settings TLSSettings) {
	handler, err := httphandlers.NewHandler(s.Auth, s.Provider)
	if err != nil {
		log.Panic(err)
	}

	if err := http.ListenAndServeTLS(settings.Addr, settings.CertFile, settings.KeyFile, handler); err != nil {
		log.Panic(err)
	}
}

func (s *Server) ReloadConfig() (err error) {
	s.Auth, err = s.Config.LoadAuth()
	if err != nil {
		return err
	}

	s.Provider, err = s.Config.LoadProvider()
	if err != nil {
		return err
	}

	return nil
}

func New(config ConfigFiles) *Server {

	p, err := config.LoadProvider()
	if err != nil {
		log.Panic(err)
	}

	a, err := config.LoadAuth()
	if err != nil {
		log.Panic(err)
	}

	return &Server{
		Config:   config,
		Auth:     a,
		Provider: p,
	}
}
