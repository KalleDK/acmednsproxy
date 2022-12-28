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
	"github.com/KalleDK/go-fpr/fpr"
	"github.com/adrg/xdg"

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
	Run: func(cmd *cobra.Command, args []string) {
		providerFile, err := getFile(providerFile, "acmednsproxy/providers.yaml")
		if err != nil {
			log.Fatal(err)
		}

		authFile, err := getFile(authFile, "acmednsproxy/auth.yaml")
		if err != nil {
			log.Fatal(err)
		}

		config := acmeserver.ConfigFiles{
			DNSType:  providers.DNSProviderName("multi"),
			DNSPath:  fpr.Resolve(providerFile),
			AuthType: acmeserver.Authenticator("simpleauth"),
			AuthPath: fpr.Resolve(authFile),
		}

		s := acmeserver.New(config)

		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGHUP)
		signal.Notify(c, syscall.SIGHUP)
		go func() {
			for range c {
				log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
				s.ReloadConfig()
			}
		}()
		log.Print("Starting server...")

		if len(certFile) > 0 {
			if len(listenAddr) == 0 {
				listenAddr = DefaultTLSAddr
			}
			log.Printf("TLS at %s\n", listenAddr)
			s.ServeTLS(acmeserver.TLSSettings{
				Addr:     listenAddr,
				CertFile: fpr.Resolve(certFile),
				KeyFile:  fpr.Resolve(keyFile),
			})
		} else {
			if len(listenAddr) == 0 {
				listenAddr = DefaultAddr
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
	serveCmd.PersistentFlags().StringVarP(&authFile, "auth", "a", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&providerFile, "providers", "p", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
