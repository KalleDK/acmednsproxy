package simpleauth

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/auth"
)

const valid_yaml = `---
type: simpleauth
permissions:
  example.com:
    user: $2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C
`

const invalid_type_yaml = `---
type: invalidtype
permissions:
  example.com:
    user: $2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C
`

const invalid_value_yaml = `---
type: noauth
permissions: 4
`

func TestLoadFromStream(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantA   *Authenticator
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				r: bytes.NewBufferString(valid_yaml),
			},
			wantA: &Authenticator{
				Permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			args: args{
				r: bytes.NewBufferString(invalid_type_yaml),
			},
			wantA:   nil,
			wantErr: true,
		},
		{
			name: "invalid value",
			args: args{
				r: bytes.NewBufferString(invalid_value_yaml),
			},
			wantA:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, err := LoadFromStream(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("LoadFromStream() = %v, want %v", gotA, tt.wantA)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantA   *Authenticator
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				r: bytes.NewBufferString(valid_yaml),
			},
			wantA: &Authenticator{
				Permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			{
				f, err := os.CreateTemp("", "")
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				io.Copy(f, tt.args.r)
				path = f.Name()

			}
			defer os.Remove(path)

			gotA, err := LoadFromFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("LoadFromFile() = %v, want %v", gotA, tt.wantA)
			}
		})
	}
}

func TestLoadFromFileError(t *testing.T) {

	t.Run("invalid file", func(t *testing.T) {
		var path string
		{
			f, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatal(err)
			}
			f.Close()
			path = f.Name()
			os.Remove(f.Name())
		}

		_, err := LoadFromFile(path)
		if err == nil {
			t.Errorf("LoadFromFile() error = %v, wantErr true", err)
			return
		}
	})
}

func TestAuthLoadFromStream(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantA   auth.Authenticator
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				r: bytes.NewBufferString(valid_yaml),
			},
			wantA: &Authenticator{
				Permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			args: args{
				r: bytes.NewBufferString(invalid_type_yaml),
			},
			wantA:   nil,
			wantErr: true,
		},
		{
			name: "invalid value",
			args: args{
				r: bytes.NewBufferString(invalid_value_yaml),
			},
			wantA:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, err := SimpleAuth.LoadFromStream(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromStream() error = %v, wantErr %v, got %v", err, tt.wantErr, gotA)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(gotA, tt.wantA) {
				t.Errorf("LoadFromStream() = %v, want %v", gotA, tt.wantA)
			}
		})
	}
}

func TestAuthenticator_Shutdown(t *testing.T) {

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string

		args    args
		wantErr bool
	}{
		{
			name: "basic",

			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{}
			if err := a.Shutdown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Authenticator.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticator_VerifyPermissions(t *testing.T) {
	type fields struct {
		permissions PermissionTable
	}
	type args struct {
		cred   auth.Credentials
		domain string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				cred:   auth.Credentials{Username: "user", Password: "password"},
				domain: "example.com",
			},
			fields: fields{
				permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "basic",
			args: args{
				cred:   auth.Credentials{Username: "user", Password: "password"},
				domain: "example.missing",
			},
			fields: fields{
				permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{
				Permissions: tt.fields.permissions,
			}
			if err := a.VerifyPermissions(tt.args.cred, tt.args.domain); (err != nil) != tt.wantErr {
				t.Errorf("Authenticator.VerifyPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticator_Save(t *testing.T) {
	type fields struct {
		Permissions PermissionTable
	}
	tests := []struct {
		name    string
		fields  fields
		want    Config
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				Permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			want: Config{
				Permissions: PermissionTable{
					"example.com": UserTable{
						"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{
				Permissions: tt.fields.Permissions,
			}
			got, err := a.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticator.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Authenticator.Save() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddPermissions(t *testing.T) {

	t.Run("add permissions", func(t *testing.T) {
		config := Config{
			Permissions: PermissionTable{
				"example.com": UserTable{
					"user": "$2a$12$gU6OnA.NwzwpgeDohXTf4.jhs9AfzcloBMYWS8nK4EhjjfNbv9E7C",
				},
			},
		}
		valid_cred := auth.Credentials{Username: "user", Password: "password"}
		empty_pass_cred := auth.Credentials{Username: "user", Password: ""}
		invalid_cred := auth.Credentials{Username: "user", Password: "password123"}
		a, err := New(config)
		if err != nil {
			t.Errorf("New() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.com")
		if err != nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.org")
		if err == nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr true", err)
		}
		err = a.VerifyPermissions(invalid_cred, "example.com")
		if err == nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr true", err)
		}
		err = a.AddPermission(valid_cred, "example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.org")
		if err != nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr false", err)
		}
		err = a.RemovePermission("user", "example.org")
		if err != nil {
			t.Errorf("RemovePermission() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.org")
		if err == nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr true", err)
		}
		err = a.RemovePermission("user", "example.org")
		if err == nil {
			t.Errorf("RemovePermission() error = %v, wantErr true", err)
		}
		err = a.AddPermission(valid_cred, "example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}
		err = a.AddPermission(valid_cred, "example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}
		err = a.AddPermission(empty_pass_cred, "example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}

	})
}
