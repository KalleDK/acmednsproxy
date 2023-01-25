/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeserver"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"

	"github.com/spf13/cobra"
)

var (
	configFile string
)

const (
	DefaultAddr    = ":8080"
	DefaultTLSAddr = ":9090"
)

type Config struct {
	ListenAddr string
	CertFile   string
	KeyFile    string
}

type Server struct {
	Server  *http.Server
	Service *acmeservice.DNSProxy
	Config  Config
}

func (s *Server) Close() {
	s.Server.Shutdown(context.Background())
}

func (s *Server) Serve() (err error) {
	if len(s.Config.CertFile) > 0 {
		addr := DefaultTLSAddr
		if s.Config.ListenAddr != "" {
			addr = s.Config.ListenAddr
		}
		s.Server, err = acmeserver.Server(s.Service, addr)
		if err != nil {
			return err
		}
		return s.Server.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile)
	}
	addr := DefaultAddr
	if s.Config.ListenAddr != "" {
		addr = s.Config.ListenAddr
	}
	s.Server, err = acmeserver.Server(s.Service, addr)
	if err != nil {
		return err
	}
	return s.Server.ListenAndServe()

}

func getFile(param, search string) (string, error) {
	if param != "" {
		return param, nil
	}
	param, err := xdg.SearchConfigFile(search)
	if err != nil {
		return "", err
	}
	return param, nil
}

func loadServer() (*Server, error) {
	configFile, err := getFile(configFile, "acmednsproxy/config.yaml")
	if err != nil {
		return nil, err
	}

	config_str, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(config_str, &config); err != nil {
		return nil, err
	}

	service, err := acmeservice.NewFromFile(configFile)
	if err != nil {
		return nil, err
	}

	return &Server{Config: config, Service: service}, nil
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var server *Server

		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGHUP)
		signal.Notify(c, syscall.SIGHUP)
		go func() {
			for range c {
				log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
				new_server, err := loadServer()
				if err != nil {
					log.Printf("Error reloading %s\n", err)
					return
				}
				server.Close()
				server = new_server
				server.Serve()
			}
		}()
		log.Print("Starting server...")

		server, err = loadServer()
		if err != nil {
			return err
		}

		if err := server.Serve(); err != nil {
			log.Print(err)
		}

		return nil

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "A help for foo")

}
