package main

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func ReverseString(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

func SortDomains(s []string) {
	for i := range s {
		s[i] = ReverseString(s[i])
	}
	sort.Strings(s)
	for i := range s {
		s[i] = ReverseString(s[i])
	}
	for i := range s[:len(s)/2] {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
}

func Decode() {
	data := `
- type: cloudflare
  demo: value
  fds: 123
- type: httpreq
  demo: value2
`

	var raw []yaml.Node
	if err := yaml.Unmarshal([]byte(data), &raw); err != nil {
		panic(err)
	}
	for _, r := range raw {
		var data map[string]interface{}
		r.Decode(&data)
		fmt.Printf("%#v\n", data)

	}

	fmt.Printf("%#v\n", raw)
}

func main() {
	data := []string{"kalle.dk.", "sort.dk.", "ns01.kalle.dk.", "aa.kalle.dk.", "dk."}
	fmt.Println(data)
	SortDomains(data)
	fmt.Println(data)
	Decode()
}
