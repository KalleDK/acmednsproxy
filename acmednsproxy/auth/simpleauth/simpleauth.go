package simpleauth

import (
	"io"
	"os"
	"path"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

const SimpleAuth = auth.Type("simpleauth")

type UserTable map[string]string

type PermissionTable map[string]UserTable

type SimpleUserAuthenticator struct {
	Permissions PermissionTable
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

func (a *SimpleUserAuthenticator) Close() (err error) { return nil }

func FromFile(path string) (*SimpleUserAuthenticator, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	var auth SimpleUserAuthenticator
	if err := auth.Load(fp); err != nil {
		return nil, err
	}

	return &auth, nil
}

func ToFile(a *SimpleUserAuthenticator, path string) (err error) {
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		close_err := fp.Close()
		if err == nil && close_err != nil {
			err = close_err
		}
	}()

	if err := a.Save(fp); err != nil {
		return err
	}

	return nil

}

type Config struct {
	Path string
}

func FromConfig(config Config) (*SimpleUserAuthenticator, error) {
	return FromFile(config.Path)
}

func Loader(unmarshal auth.YAMLUnmarshaler, config_dir string) (auth.Authenticator, error) {
	var conf Config
	if err := unmarshal(&conf); err != nil {
		return nil, err
	}

	if !path.IsAbs(conf.Path) {
		conf.Path = path.Join(config_dir, conf.Path)
	}

	return FromConfig(conf)
}

func init() {
	SimpleAuth.Register(Loader)
}
