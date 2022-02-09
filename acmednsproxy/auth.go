package acmednsproxy

import (
	"encoding/json"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrEmptyFile     = errors.New("empty file")
	ErrUnknownDomain = errors.New("unknown domain")
	ErrUnknownUser   = errors.New("unknown user")
)

type UserAuthenticator interface {
	VerifyPermissions(user string, password string, domain string) (err error)
}

type UserTable map[string][]byte

type PermissionTable map[string]UserTable

type SimpleUserAuthenticator struct {
	Permissions PermissionTable
}

func (a *SimpleUserAuthenticator) Load(f io.Reader) (err error) {
	dec := json.NewDecoder(f)
	if err := dec.Decode(&a.Permissions); err != nil {
		return err
	}
	return nil
}

func (a *SimpleUserAuthenticator) Save(w io.Writer) (err error) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a.Permissions); err != nil {
		return err
	}
	return nil
}

func (a *SimpleUserAuthenticator) AddPermission(user string, password string, domain string) (err error) {
	if a.Permissions == nil {
		a.Permissions = PermissionTable{}
	}

	encodedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users, ok := a.Permissions[domain]
	if !ok {
		users = UserTable{}
		a.Permissions[domain] = users
	}

	users[user] = encodedPassword
	return nil
}

func (a *SimpleUserAuthenticator) RemovePermission(user string, domain string) (err error) {
	if a.Permissions == nil {
		return ErrEmptyFile
	}

	users, ok := a.Permissions[domain]
	if !ok {
		return ErrUnknownDomain
	}

	_, ok = users[user]
	if !ok {
		return ErrUnknownUser
	}

	delete(users, user)
	return nil
}

func (a *SimpleUserAuthenticator) VerifyPermissions(user string, password string, domain string) (err error) {
	users, ok := a.Permissions[domain]
	if !ok {
		return ErrUnauthorized
	}

	encodedPassword, ok := users[user]
	if !ok {
		return ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword(encodedPassword, []byte(password)); err != nil {
		return ErrUnauthorized
	}

	return nil
}
