package cmd

import (
	"fmt"
	"syscall"

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
