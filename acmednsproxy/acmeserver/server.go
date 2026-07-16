package acmeserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
)

// Server bundles together all the components needed to run a single instance
// of the acmednsproxy HTTP server.
type Server struct {
	TLS        *TLSService
	Proxy      *acmeservice.DNSProxy
	HTTPServer *http.Server
	Config     Config
}

// Reload refreshes the TLS certificate and the DNS proxy configuration
// (authenticator + providers) from disk without restarting the server.
func (s *Server) Reload() error {
	if err := s.TLS.Reload(); err != nil {
		return err
	}

	return s.Proxy.Reload()
}

// Shutdown gracefully stops the HTTP server using ctx to bound the wait.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}

// Close immediately closes the HTTP server.
func (s *Server) Close() error {
	return s.HTTPServer.Close()
}

// ServeTLS starts the HTTPS server.  The TLS certificate is served via
// TLSService.GetCertificate so it can be rotated at runtime via /reload.
func (s *Server) ServeTLS() error {
	handler, err := NewHandler(s.Proxy, s.TLS)
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
	}
	fmt.Println("Listening on", s.Config.Listen)
	server := &http.Server{
		Addr:    s.Config.Listen,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: s.TLS.GetCertificate,
		},
	}
	return server.ListenAndServeTLS("", "")
}

// Serve starts the plain-HTTP server.
func (s *Server) Serve() error {
	handler, err := NewHandler(s.Proxy, s.TLS)
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
	}

	return http.ListenAndServe(s.Config.Listen, handler)
}

// ListenAndServe starts either a plain-HTTP or HTTPS server depending on
// whether TLS has been configured.
func (s *Server) ListenAndServe() error {
	if s.TLS == nil {
		fmt.Println("No TLS Configured")
		return s.Serve()
	}
	fmt.Println("TLS Configured")
	return s.ServeTLS()
}

// loadServer reads the config file at path and constructs a Server.
func loadServer(path string) (server Server, err error) {
	config, err := loadConfig(path)
	if err != nil {
		return
	}

	var tlsSvc *TLSService
	if config.HasTLS() {
		tlsSvc, err = NewTLSService(config.TLS)
		if err != nil {
			return
		}
	}

	service, err := acmeservice.New(config.Proxy)
	if err != nil {
		return
	}

	return Server{
		HTTPServer: nil,
		Config:     config,
		TLS:        tlsSvc,
		Proxy:      service,
	}, nil
}

// ServerWithConfig wraps Server and supports hot-reloading the full config
// (including the listen address and TLS settings) by replacing the inner
// Server on each Reload call.
type ServerWithConfig struct {
	// ConfigFile is the path to the main YAML config file.
	ConfigFile string
	// IsClosing is set to true when Shutdown or Close is called so that the
	// ListenAndServe loop exits cleanly.
	IsClosing bool
	Server
}

// Reload loads a fresh Server from ConfigFile, swaps it with the current one,
// and shuts down the old server so the new one can start listening.
func (s *ServerWithConfig) Reload(ctx context.Context) (err error) {
	if s.IsClosing {
		return http.ErrServerClosed
	}
	var newServer, oldServer Server

	if newServer, err = loadServer(s.ConfigFile); err != nil {
		return err
	}

	oldServer, s.Server = s.Server, newServer

	oldServer.Shutdown(ctx)

	return nil
}

// Close marks the server as closing and immediately closes the inner HTTP
// server.
func (s *ServerWithConfig) Close() error {
	s.IsClosing = true
	return s.Server.Close()
}

// Shutdown marks the server as closing and gracefully shuts down the inner
// HTTP server.
func (s *ServerWithConfig) Shutdown(ctx context.Context) error {
	s.IsClosing = true
	return s.Server.Shutdown(ctx)
}

// ListenAndServe continuously calls the inner Server's ListenAndServe,
// restarting after each Reload until IsClosing is set.
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

// NewServer constructs a ServerWithConfig by loading the config file at path.
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
