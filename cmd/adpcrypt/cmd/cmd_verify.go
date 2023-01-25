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

var verifyFlags = &struct {
	User     string
	Domain   string
	AuthFile string
	Password string
	AskPass  bool
}{}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:                   "verify [-d domain] [-u user] {-k | -p password} [-a file]",
	Short:                 "Verify username and password agains domain",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		flags := verifyFlags

		s, err := loadAuthFile(flags.AuthFile)
		if err != nil {
			return err
		}

		if flags.Password == "" && !flags.AskPass {
			return fmt.Errorf("either give password with arg or use -p")
		}

		for flags.Domain == "" {
			fmt.Print("Enter Domain: ")
			fmt.Scanln(&flags.Domain)
		}

		for flags.User == "" {
			fmt.Print("Enter Username: ")
			fmt.Scanln(&flags.User)
		}

		if flags.AskPass {
			flags.Password, err = getPassword()
			if err != nil {
				return err
			}
		}

		if err := s.VerifyPermissions(flags.User, flags.Password, flags.Domain); err != nil {
			fmt.Println("Not Ok!")
		} else {
			fmt.Println("Ok!")
		}

		return nil
	},
}

func init() {
	cmd := verifyCmd
	flags := verifyFlags

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

	flagname = "password"
	cmd.Flags().StringVarP(&flags.Password, flagname, "p", "", "password")
	cmd.RegisterFlagCompletionFunc(flagname, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	flagname = "ask-pass"
	cmd.Flags().BoolVarP(&flags.AskPass, flagname, "k", false, "Ask for a password")

	flagname = "auth"
	cmd.Flags().StringVarP(&flags.AuthFile, flagname, "a", defaultAuthFile, "auth file")
	cmd.MarkFlagFilename(flagname, "yaml", "yml")

	rootCmd.AddCommand(cmd)
}
