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
	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"

	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/auth/all"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"

	"github.com/spf13/cobra"
)

var (
	authFile     string
	providerFile string
	certFile     string
	keyFile      string
	listenAddr   string
)

const (
	defaultAddr         = ":8080"
	defaultTLSAddr      = ":9090"
	defaultAuthFile     = "auth.yaml"
	defaultProviderFile = "providers.yaml"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGHUP)
		signal.Notify(c, syscall.SIGHUP)

		config := acmeserver.ConfigFiles{
			DNSType:  providers.DNSProviderName("multi"),
			DNSPath:  providerFile,
			AuthType: acmeserver.Authenticator("simpleauth"),
			AuthPath: authFile,
		}

		s := acmeserver.New(config)

		go func() {
			for range c {
				log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
				s.ReloadConfig()
			}
		}()
		log.Print("Starting server...")

		if len(certFile) > 0 {
			if len(listenAddr) == 0 {
				listenAddr = defaultTLSAddr
			}
			log.Printf("TLS at %s\n", listenAddr)
			s.ServeTLS(acmeserver.TLSSettings{
				Addr:     listenAddr,
				CertFile: certFile,
				KeyFile:  keyFile,
			})
		} else {
			if len(listenAddr) == 0 {
				listenAddr = defaultAddr
			}
			log.Printf("Non-TLS at %s\n", listenAddr)
			s.Serve(acmeserver.Settings{
				Addr: listenAddr,
			})
		}

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	serveCmd.PersistentFlags().StringVarP(&certFile, "cert", "c", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&keyFile, "key", "k", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&listenAddr, "addr", "i", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&authFile, "auth", "a", defaultAuthFile, "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&providerFile, "providers", "p", defaultProviderFile, "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
