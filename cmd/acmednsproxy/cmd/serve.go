/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeserver"
	"github.com/adrg/xdg"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"

	"github.com/spf13/cobra"
)

var (
	configFile string
)

func findFile(param, search string) (string, error) {
	if param != "" {
		return param, nil
	}
	param, err := xdg.SearchConfigFile(search)
	if err != nil {
		return "", err
	}
	return param, nil
}

func loadServer() (*acmeserver.ServerWithConfig, error) {
	configFile, err := findFile(configFile, "acmednsproxy/config.yaml")
	if err != nil {
		return nil, err
	}

	server, err := acmeserver.NewServer(configFile)
	if err != nil {
		return nil, err
	}

	return server, err
}

func reloadOn(server *acmeserver.ServerWithConfig, ev os.Signal, ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGHUP)
	signal.Notify(c, syscall.SIGHUP)
	for {
		select {
		case <-c:
			log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
			server.Reload(context.Background())
		case <-ctx.Done():
			return
		}
	}
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		server, err := loadServer()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())

		go reloadOn(server, syscall.SIGHUP, ctx)

		log.Print("Starting server...")

		if err := server.ListenAndServe(); err != nil {
			log.Print(err)
		}

		cancel()

		return nil

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "A help for foo")
}
