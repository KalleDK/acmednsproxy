package auth

import (
	"errors"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrUnknownDomain = errors.New("unknown domain")
	ErrUnknownUser   = errors.New("unknown user")
)

type UserAuthenticator interface {
	VerifyPermissions(user string, password string, domain string) (err error)
}
