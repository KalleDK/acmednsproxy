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
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth/simpleauth"
	"github.com/spf13/cobra"
)

var initFlags = &struct {
	AuthFile string
}{}

// addCmd represents the add command
var initCmd = &cobra.Command{
	Use:                   "init [-a file]",
	Short:                 "Create an empty auth file",
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		flags := initFlags
		s := &simpleauth.SimpleUserAuthenticator{
			Permissions: simpleauth.PermissionTable{},
		}

		if err := saveAuthFile(flags.AuthFile, s); err != nil {
			return err
		}

		return nil

	},
}

func init() {
	cmd := initCmd
	flags := initFlags

	flagname := "auth"
	cmd.Flags().StringVarP(&flags.AuthFile, flagname, "a", defaultAuthFile, "auth file")
	cmd.MarkFlagFilename(flagname, "yaml", "yml")

	rootCmd.AddCommand(cmd)
}
