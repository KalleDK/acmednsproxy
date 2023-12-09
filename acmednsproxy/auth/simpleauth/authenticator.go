package simpleauth

import (
	"context"
	"fmt"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	Permissions PermissionTable
}

func (a *Authenticator) AddPermission(cred auth.Credentials, domain string) (err error) {
	if a.Permissions == nil {
		a.Permissions = PermissionTable{}
	}

	encodedPassword, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users, ok := a.Permissions[domain]
	if !ok {
		users = UserTable{}
		a.Permissions[domain] = users
	}

	users[cred.Username] = string(encodedPassword)
	return nil
}

func (a *Authenticator) RemovePermission(user string, domain string) (err error) {
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

func (a *Authenticator) VerifyPermissions(cred auth.Credentials, domain string) (err error) {
	users, ok := a.Permissions[domain]
	if !ok {
		return fmt.Errorf("domain does not exists in auth %s %w", domain, auth.ErrUnauthorized)
	}

	encodedPassword, ok := users[cred.Username]
	if !ok {
		return fmt.Errorf("user does not exists in domain %s %s %w", cred.Username, domain, auth.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(cred.Password)); err != nil {
		return fmt.Errorf("password does not match %s %s %w", cred.Username, domain, auth.ErrUnauthorized)
	}

	return nil
}

func (a *Authenticator) Load(config Config) (err error) {
	a.Permissions = PermissionTable{}
	for domain, users := range config.Permissions {
		a.Permissions[domain] = UserTable{}
		for user, password := range users {
			a.Permissions[domain][user] = password
		}
	}

	return nil
}

func (a *Authenticator) Save() (Config, error) {
	config := Config{}
	config.Permissions = PermissionTable{}
	for domain, users := range a.Permissions {
		config.Permissions[domain] = UserTable{}
		for user, password := range users {
			config.Permissions[domain][user] = password
		}
	}
	return config, nil
}

func (a *Authenticator) Close() (err error) { return nil }

func (a *Authenticator) Shutdown(ctx context.Context) error {
	return a.Close()
}

func New(config Config) (a *Authenticator, err error) {
	a = &Authenticator{}
	a.Load(config)
	return a, nil
}
