package simpleauth

import (
	"io"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/acmeserver"
	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type UserTable map[string]string

type PermissionTable map[string]UserTable

type SimpleUserAuthenticator struct {
	Permissions PermissionTable
}

func (a *SimpleUserAuthenticator) Load(f io.Reader) (err error) {
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&a.Permissions); err != nil {
		return err
	}
	return nil
}

func (a *SimpleUserAuthenticator) Save(w io.Writer) (err error) {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
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

	users[user] = string(encodedPassword)
	return nil
}

func (a *SimpleUserAuthenticator) RemovePermission(user string, domain string) (err error) {
	if a.Permissions == nil {
		return auth.ErrUnknownDomain
	}

	users, ok := a.Permissions[domain]
	if !ok {
		return auth.ErrUnknownDomain
	}

	_, ok = users[user]
	if !ok {
		return auth.ErrUnknownUser
	}

	delete(users, user)
	return nil
}

func (a *SimpleUserAuthenticator) VerifyPermissions(user string, password string, domain string) (err error) {
	users, ok := a.Permissions[domain]
	if !ok {
		return auth.ErrUnauthorized
	}

	encodedPassword, ok := users[user]
	if !ok {
		return auth.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(password)); err != nil {
		return auth.ErrUnauthorized
	}

	return nil
}

type SimpleUserAuthenticatorLoader struct {
	Path string
}

func (a SimpleUserAuthenticatorLoader) Load(d acmeserver.ConfigDecoder) (uauth auth.Authenticator, err error) {

	u := SimpleUserAuthenticator{}
	if err := d.Decode(&u.Permissions); err != nil {
		return nil, err
	}

	return &u, nil
}

func init() {
	acmeserver.RegisterAuthenticator(acmeserver.Authenticator("simpleauth"), SimpleUserAuthenticatorLoader{})
}
