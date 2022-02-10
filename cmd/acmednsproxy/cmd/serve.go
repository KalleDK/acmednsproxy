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
	"github.com/spf13/cobra"
)

var authFile string
var providerFile string
var certFile string
var keyFile string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGHUP)
		signal.Notify(c, syscall.SIGHUP)
		s := acmeserver.New(authFile, providerFile, certFile, keyFile)

		go func() {
			for range c {
				log.Printf("Got A HUP Signal! Now Reloading Conf....\n")
				s.ReloadConfig()
			}
		}()
		log.Print("Starting server")
		s.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	serveCmd.PersistentFlags().StringVarP(&certFile, "cert", "c", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&keyFile, "key", "k", "", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&authFile, "auth", "a", "auth.json", "A help for foo")
	serveCmd.PersistentFlags().StringVarP(&providerFile, "providers", "p", "providers.json", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
