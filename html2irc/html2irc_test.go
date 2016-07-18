package html2irc

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
	{"<b><a href='what'><div>ever</div></a></b>", []string{"\x02[ever](what)"}},
	{"a<br><br>b", []string{"a", "b"}},
	{"<button value='about:blank'>A button!</button>", []string{"[A button!](about:blank)"}},
}

func TestHTMLToIRC(t *testing.T) {
	for _, test := range tests {
		out := HTMLToIRC(test.in)
		if !reflect.DeepEqual(out, test.out) {
			t.Errorf("HTMLToIRC(%q) => %q, want %q", test.in, out, test.out)
		}
	}
}
