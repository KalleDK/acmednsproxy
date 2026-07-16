package cloudflare

import (
	"reflect"
	"testing"
)

func TestReverseString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "single char", input: "a", want: "a"},
		{name: "ascii", input: "hello", want: "olleh"},
		{name: "domain", input: "example.com.", want: ".moc.elpmaxe"},
		{name: "unicode", input: "héllo", want: "olléh"},
		{name: "idempotent on palindrome", input: "racecar", want: "racecar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseString(tt.input)
			if got != tt.want {
				t.Errorf("ReverseString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReverseStringIdempotent(t *testing.T) {
	inputs := []string{"", "a", "hello", "example.com.", "sub.example.com."}
	for _, s := range inputs {
		if got := ReverseString(ReverseString(s)); got != s {
			t.Errorf("ReverseString(ReverseString(%q)) = %q, want %q", s, got, s)
		}
	}
}

func TestSortDomains(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "empty",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "single",
			input: []string{"example.com."},
			want:  []string{"example.com."},
		},
		{
			name:  "sub before apex",
			input: []string{"example.com.", "sub.example.com."},
			want:  []string{"sub.example.com.", "example.com."},
		},
		{
			name:  "already most-specific first",
			input: []string{"sub.example.com.", "example.com."},
			want:  []string{"sub.example.com.", "example.com."},
		},
		{
			// The algorithm reverses each string, sorts lexicographically,
			// reverses back, then reverses the whole slice.  The net effect
			// is most-specific-first within each subtree, but the relative
			// ordering of unrelated same-level domains follows the reversed-
			// string sort.  The expected value here is the actual output of
			// the algorithm.
			name:  "multiple subdomains",
			input: []string{"example.com.", "a.example.com.", "b.example.com.", "c.a.example.com."},
			want:  []string{"b.example.com.", "c.a.example.com.", "a.example.com.", "example.com."},
		},
		{
			// Unrelated domains (no parent-child relationship) are ordered by
			// the reversed-string sort descending.
			name:  "different TLDs",
			input: []string{"example.org.", "example.com."},
			want:  []string{"example.com.", "example.org."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]string, len(tt.input))
			copy(got, tt.input)
			SortDomains(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortDomains(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestSortDomainsLongestMatchFirst verifies the core invariant: for any pair
// of domains in the sorted slice where one is a subdomain of the other, the
// subdomain must appear before the parent.
func TestSortDomainsLongestMatchFirst(t *testing.T) {
	input := []string{
		"example.com.",
		"deep.sub.example.com.",
		"sub.example.com.",
		"other.example.com.",
	}
	SortDomains(input)

	indexOf := func(s []string, v string) int {
		for i, x := range s {
			if x == v {
				return i
			}
		}
		return -1
	}

	pairs := [][2]string{
		{"deep.sub.example.com.", "sub.example.com."},
		{"deep.sub.example.com.", "example.com."},
		{"sub.example.com.", "example.com."},
		{"other.example.com.", "example.com."},
	}

	for _, p := range pairs {
		child, parent := p[0], p[1]
		ci, pi := indexOf(input, child), indexOf(input, parent)
		if ci == -1 || pi == -1 {
			t.Fatalf("domain missing from sorted slice: %v", input)
		}
		if ci > pi {
			t.Errorf("expected %q (idx %d) before %q (idx %d) in %v", child, ci, parent, pi, input)
		}
	}
}
