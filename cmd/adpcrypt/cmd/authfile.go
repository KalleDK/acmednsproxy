package cmd

import (
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth/simpleauth"
)

func loadAuthFile(path string) (a *simpleauth.SimpleUserAuthenticator, err error) {
	return simpleauth.FromFile(path)
}

func saveAuthFile(path string, a *simpleauth.SimpleUserAuthenticator) (err error) {
	return simpleauth.ToFile(a, path)
}
