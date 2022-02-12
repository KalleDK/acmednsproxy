package cmd

import (
	"fmt"
	"syscall"

	"github.com/sethvargo/go-password/password"
	"golang.org/x/term"
)

func getPassword() (pass string, err error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return
	}
	fmt.Println("")
	return string(bytePassword), nil
}

func generatePassword() (pass string, err error) {
	pass, err = password.Generate(24, 5, 0, false, false)
	if err != nil {
		return
	}

	return pass, nil
}
