/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const output = `
HTTPREQ_MODE=RAW \
HTTPREQ_USERNAME=%s \
HTTPREQ_PASSWORD=%s

`

var addFlags = &struct {
	User     string
	Domain   string
	AuthFile string
	AskPass  bool
}{}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:                   "add [-d domain] [-u user] [-a file] [-k]",
	Short:                 "Add user to domain",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		flags := addFlags

		s, err := loadAuthFile(flags.AuthFile)
		if err != nil {
			return err
		}

		passFunc := generatePassword
		if flags.AskPass {
			passFunc = getPassword
		}

		for flags.Domain == "" {
			fmt.Print("Enter Domain: ")
			fmt.Scanln(&flags.Domain)
		}

		for flags.User == "" {
			fmt.Print("Enter Username: ")
			fmt.Scanln(&flags.User)
		}

		password, err := passFunc()
		if err != nil {
			return err
		}

		if err := s.AddPermission(flags.User, password, flags.Domain); err != nil {
			return err
		}

		if err := saveAuthFile(flags.AuthFile, s); err != nil {
			return err
		}

		fmt.Printf(output, flags.User, password)

		return nil

	},
}

func init() {

	flags := addFlags
	cmd := addCmd

	flagname := "domain"
	cmd.Flags().StringVarP(&flags.Domain, flagname, "d", "", "domain")
	cmd.RegisterFlagCompletionFunc(flagname, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	flagname = "user"
	cmd.Flags().StringVarP(&flags.User, flagname, "u", "", "username")
	cmd.RegisterFlagCompletionFunc(flagname, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	flagname = "auth"
	cmd.Flags().StringVarP(&flags.AuthFile, flagname, "a", defaultAuthFile, "auth file")
	cmd.MarkFlagFilename(flagname, "yaml", "yml")

	flagname = "ask-pass"
	cmd.Flags().BoolVarP(&flags.AskPass, flagname, "k", false, "Ask for a password")

	rootCmd.AddCommand(cmd)
}
