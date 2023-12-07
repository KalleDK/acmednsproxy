package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"

	"github.com/KalleDK/acmednsproxy/acmednsproxy/providers"
	_ "github.com/KalleDK/acmednsproxy/acmednsproxy/providers/all"
	"gopkg.in/yaml.v3"
)

//go:embed multi.yml
var conf_str string

func main() {

	p, err := providers.LoadFromStream(bytes.NewBufferString(conf_str))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", p)

	var a any
	err = yaml.Unmarshal([]byte("---\n"), &a)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", a)
}
