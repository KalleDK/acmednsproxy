package acmeserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
)

type Server struct {
	TLS        *TLSService
	Proxy      *acmeservice.DNSProxy
	HTTPServer *http.Server
	Config     Config
}

func (s *Server) Reload() error {
	if err := s.TLS.Reload(); err != nil {
		return err
	}

	return s.Proxy.Reload()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.HTTPServer.Close()
}

func (s *Server) ServeTLS() error {
	handler, err := NewHandler(s.Proxy, s.TLS)
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
	}
	fmt.Println("Listening on", s.Config.Listen)
	server := &http.Server{Addr: s.Config.Listen, Handler: handler, TLSConfig: &tls.Config{GetCertificate: s.TLS.GetCertificate}}
	return server.ListenAndServeTLS("", "")
}

func (s *Server) Serve() error {
	handler, err := NewHandler(s.Proxy, s.TLS)
	if err != nil {
		log.Panic(err)
	}

	return http.ListenAndServe(s.Config.Listen, handler)
}

func (s *Server) ListenAndServe() error {
	if s.TLS == nil {
		fmt.Println("No TLS Configured")
		return s.Serve()
	}
	fmt.Println("TLS Configured")
	return s.ServeTLS()
}

func loadServer(path string) (server Server, err error) {
	config, err := loadConfig(path)
	if err != nil {
		return
	}

	var tls *TLSService
	if config.HasTLS() {
		tls, err = NewTLSService(config.TLS)
		if err != nil {
			return
		}
	} else {
		tls = nil
	}

	service, err := acmeservice.New(config.Proxy)
	if err != nil {
		return
	}

	return Server{
		HTTPServer: nil,
		Config:     config,
		TLS:        tls,
		Proxy:      service,
	}, nil
}

type ServerWithConfig struct {
	ConfigFile string
	IsClosing  bool
	Server
}

func (s *ServerWithConfig) Reload(ctx context.Context) (err error) {
	if s.IsClosing {
		return http.ErrServerClosed
	}
	var new_server, old_server Server

	if new_server, err = loadServer(s.ConfigFile); err != nil {
		return err
	}

	old_server, s.Server = s.Server, new_server

	old_server.Shutdown(ctx)

	return nil
}

func (s *ServerWithConfig) Close() error {
	s.IsClosing = true
	return s.Server.Close()
}

func (s *ServerWithConfig) Shutdown(ctx context.Context) error {
	s.IsClosing = true
	return s.Server.Shutdown(ctx)
}

func (s *ServerWithConfig) ListenAndServe() error {
	for {
		if s.IsClosing {
			return http.ErrServerClosed
		}

		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
	}
}

func NewServer(path string) (*ServerWithConfig, error) {
	server, err := loadServer(path)
	if err != nil {
		return nil, err
	}

	return &ServerWithConfig{
		ConfigFile: path,
		IsClosing:  false,
		Server:     server,
	}, nil
}
