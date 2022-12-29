package main

import (
	"fmt"
	"sort"
	"unicode/utf8"
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

func main() {
	data := []string{"kalle.dk.", "sort.dk.", "ns01.kalle.dk.", "aa.kalle.dk.", "dk."}
	fmt.Println(data)
	SortDomains(data)
	fmt.Println(data)
}
