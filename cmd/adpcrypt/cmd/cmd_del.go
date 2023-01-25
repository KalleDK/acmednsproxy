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

var delFlags = &struct {
	User     string
	Domain   string
	AuthFile string
}{}

// delCmd represents the del command
var delCmd = &cobra.Command{
	Use:                   "del [-d domain] [-u user] [-a file]",
	Short:                 "Remove user from domain",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		flags := delFlags

		s, err := loadAuthFile(flags.AuthFile)
		if err != nil {
			return err
		}

		for flags.Domain == "" {
			fmt.Print("Enter Domain: ")
			fmt.Scanln(&flags.Domain)
		}

		for flags.User == "" {
			fmt.Print("Enter Username: ")
			fmt.Scanln(&flags.User)
		}

		if err := s.RemovePermission(flags.User, flags.Domain); err != nil {
			return err
		}

		if err := saveAuthFile(flags.AuthFile, s); err != nil {
			return err
		}

		return nil

	},
}

func init() {

	cmd := delCmd
	flags := delFlags

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

	rootCmd.AddCommand(cmd)
}
