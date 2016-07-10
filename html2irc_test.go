package main

import (
	"reflect"
	"testing"
)

var tests = []struct {
	in  string
	out []string
}{
	{"a", []string{"a"}},
	{"<div>a</div><div>b</div>", []string{"a", "b"}},
	{"<div><div>a</div>b<span>c</span><div>d</div></div>e", []string{"a", "bc", "d", "e"}},
	{"<a href='what'><div>ever</div></a>", []string{"[ever](what)"}},
}

func TestHTMLToIRC(t *testing.T) {
	for _, test := range tests {
		out := htmlToIRC(test.in)
		if !reflect.DeepEqual(out, test.out) {
			t.Errorf("htmlToIRC(%q) => %q, want %q", test.in, out, test.out)
		}
	}
}
