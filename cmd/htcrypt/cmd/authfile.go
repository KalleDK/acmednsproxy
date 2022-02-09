package cmd

import (
	"bytes"
	"os"

	"github.com/KalleDK/acmednsproxy/acmednsproxy"
)

func loadAuthFile() (a acmednsproxy.SimpleUserAuthenticator, err error) {
	data, err := os.ReadFile(authFile)
	if err != nil {
		return
	}

	if err = a.Load(bytes.NewReader(data)); err != nil {
		return
	}

	return a, nil
}

func saveAuthFile(a acmednsproxy.SimpleUserAuthenticator) (err error) {
	data := bytes.Buffer{}

	if err = a.Save(&data); err != nil {
		return
	}

	if err = os.WriteFile(authFile, data.Bytes(), 0644); err != nil {
		return
	}

	return nil
}
