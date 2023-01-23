/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeserver"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeservice"
	"github.com/KalleDK/go-fpr/fpr"
	"github.com/adrg/xdg"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"

	"github.com/spf13/cobra"
)

var (
	configFile string
	certFile   string
	keyFile    string
	listenAddr string
)

var (
	DefaultAddr    = ":8080"
	DefaultTLSAddr = ":9090"
)

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

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, err := getFile(configFile, "acmednsproxy/config.yaml")
		if err != nil {
			log.Fatal(err)
		}

		service, err := acmeservice.NewFromFile(configFile)
		if err != nil {
			return err
		}

		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGHUP)
		signal.Notify(c, syscall.SIGHUP)
		go func() {
			for range c {
				log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
				service.Reload()
			}
		}()
		log.Print("Starting server...")

		if len(certFile) > 0 {
			if len(listenAddr) == 0 {
				listenAddr = DefaultTLSAddr
			}
			log.Printf("TLS at %s\n", listenAddr)
			acmeserver.ServeTLS(service, acmeserver.TLSSettings{
				Addr:     listenAddr,
				CertFile: fpr.Resolve(certFile),
				KeyFile:  fpr.Resolve(keyFile),
			})
		} else {
			if len(listenAddr) == 0 {
				listenAddr = DefaultAddr
			}
			log.Printf("Non-TLS at %s\n", listenAddr)
			acmeserver.Serve(service, acmeserver.Settings{
				Addr: listenAddr,
			})
		}

		return nil

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&certFile, "cert", "p", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&keyFile, "key", "k", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&listenAddr, "addr", "i", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "A help for foo")

}
