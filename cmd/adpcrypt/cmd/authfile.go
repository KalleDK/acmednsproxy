package cmd

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth/simpleauth"
)

func loadAuthFile(path string) (a *simpleauth.Authenticator, err error) {
	return simpleauth.LoadFromFile(path)
}

func saveAuthFile(path string, a *simpleauth.Authenticator) (err error) {
	return simpleauth.SaveToFile(path, a)
}
