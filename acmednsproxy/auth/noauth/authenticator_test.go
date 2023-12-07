package noauth

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
type: noauth
domains:
- example.com
- example.org
`

const invalid_type_yaml = `---
type: invalidtype
domains:
- exampleinv.com
- exampleinv.org
`

const invalid_value_yaml = `---
type: noauth
domains: 4
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
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
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
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
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
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
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
			gotA, err := NoAuth.LoadFromStream(tt.args.r)
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
	type fields struct {
		domains map[string]struct{}
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "basic",
			fields: fields{},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{
				domains: tt.fields.domains,
			}
			if err := a.Shutdown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Authenticator.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticator_VerifyPermissions(t *testing.T) {
	type fields struct {
		domains map[string]struct{}
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
				cred: auth.Credentials{Username: "user", Password: "password"},

				domain: "example.com",
			},
			fields: fields{
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
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
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{
				domains: tt.fields.domains,
			}
			if err := a.VerifyPermissions(tt.args.cred, tt.args.domain); (err != nil) != tt.wantErr {
				t.Errorf("Authenticator.VerifyPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticator_Save(t *testing.T) {
	type fields struct {
		domains map[string]struct{}
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
				domains: map[string]struct{}{
					"example.com": {},
					"example.org": {},
				},
			},
			want: Config{
				Domains: []string{"example.com", "example.org"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authenticator{
				domains: tt.fields.domains,
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
			Domains: []string{"example.com"},
		}
		valid_cred := auth.Credentials{Username: "user", Password: "password"}
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
		err = a.AddPermission("example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.org")
		if err != nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr false", err)
		}
		err = a.RemovePermission("example.org")
		if err != nil {
			t.Errorf("RemovePermission() error = %v, wantErr false", err)
		}
		err = a.VerifyPermissions(valid_cred, "example.org")
		if err == nil {
			t.Errorf("VerifyPermissions() error = %v, wantErr true", err)
		}
		err = a.RemovePermission("example.org")
		if err == nil {
			t.Errorf("RemovePermission() error = %v, wantErr true", err)
		}
		err = a.AddPermission("example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}
		err = a.AddPermission("example.org")
		if err != nil {
			t.Errorf("AddPermission() error = %v, wantErr false", err)
		}

	})
}
