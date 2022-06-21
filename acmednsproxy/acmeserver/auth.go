package acmeserver

import "github.com/KalleDK/acmednsproxy/acmednsproxy/auth"

type AuthenticatorLoader interface {
	Load(d ConfigDecoder) (a auth.Authenticator, err error)
}

type Authenticator string

var authenticatorMap = map[Authenticator]AuthenticatorLoader{}

func (u Authenticator) Load(d ConfigDecoder) (a auth.Authenticator, err error) {
	return authenticatorMap[u].Load(d)
}

func RegisterAuthenticator(name Authenticator, l AuthenticatorLoader) {
	authenticatorMap[name] = l
}
