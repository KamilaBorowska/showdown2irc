package html2irc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	{"<input type='button' value='yay'>", []string{"yay"}},
	{"<img src='http://example.com' alt='an alt'>", []string{"an alt"}},
	{"<span style='display: block'>a</span><span>b</span><span style='display: hidden'>c</span><span>d</span>", []string{"a", "bd"}},
	{"<span style='display; display:;'>a</span>", []string{"a"}},
	{"<ul><li>a<li>b<li>c</ul>", []string{"• a", "• b", "• c"}},
	{"yes<h1>Heading</h1>no", []string{"yes", "\x02Heading", "no"}},
	{"", nil},
	{strings.Repeat("<div>", 10000), nil},
}

func TestHTMLToIRC(t *testing.T) {
	for _, test := range tests {
		assert.Equal(t, HTMLToIRC(test.in), test.out, "HTMLToIRC(%#q)", test.in)
	}
}
