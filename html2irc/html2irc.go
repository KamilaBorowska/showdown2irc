// showdown2irc - use Showdown chat with an IRC client
// Copyright (C) 2016 Konrad Borowski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package html2irc

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

const boldCharacter = "\x02"

type htmlConverter struct {
	*bytes.Buffer
	*html.Tokenizer
	preventNewLines *bool
	bold, hidden    bool
}

// HTMLToIRC converts HTML code into IRC text.
//
// IRC text is different to regular text in that it can use IRC formatting
// sequences such as bold.
func HTMLToIRC(code string) []string {
	converter := htmlConverter{
		Buffer:          new(bytes.Buffer),
		Tokenizer:       html.NewTokenizer(strings.NewReader(code)),
		preventNewLines: new(bool),
	}
	converter.parseToken()
	var result []string
	for _, line := range strings.Split(converter.String(), "\n") {
		trimmedLine := strings.Replace(
			strings.TrimRight(strings.TrimSpace(line), boldCharacter), boldCharacter+boldCharacter, "", -1)

		if !isEmpty(trimmedLine) {
			result = append(result, trimmedLine)
		}
	}
	return result
}

func (c htmlConverter) parseToken() {
	for {
		tt := c.Tokenizer.Next()
		switch tt {
		case html.StartTagToken:
			c.parseStartToken()

		case html.TextToken:
			if !c.hidden {
				*c.preventNewLines = false
				c.Write(c.Text())
			}

		case html.EndTagToken, html.ErrorToken:
			return
		}
	}
}

func (c htmlConverter) printNewline() {
	if *c.preventNewLines {
		return
	}
	c.WriteByte('\n')
	if c.bold {
		c.WriteByte(boldCharacter[0])
	}
}

func (c htmlConverter) chopNewLines() string {
	bytes := c.Bytes()
	i := len(bytes) - 1
	for i >= 0 && isNewLineCharacter(bytes[i]) {
		i--
	}
	i++
	out := string(bytes[i:])
	c.Truncate(i)
	return out
}

func (c htmlConverter) writeLink(link string) {
	newLines := c.chopNewLines()
	c.WriteString(fmt.Sprintf("](%s)%s", link, newLines))
}

func (c htmlConverter) parseStartToken() {
	// This is a really basic HTML rendering engine for purpose of decoding
	// raw text (such as output from commands).
	//
	// In CSS, there are two major display categories, inline and block.
	// If an element is a block element, it means it takes its own line.
	// For instance, the following HTML code:
	//
	//     <div><div>a</div>b<span>c</span><div>d</div></div>e
	//
	// Can be rendered as this:
	//
	//     a
	//     bc
	//     d
	//     e
	//
	// This converter handles block elements by putting newlines on both
	// of sides of the element. This has side-effect of putting way too
	// many new lines which are dealt with by removing empty lines on
	// output.
	//
	// Certain elements such as <b> can be represented using IRC special
	// formatting characters. They are dealt by preserving a state that
	// is also used by recursive function calls (in order to prevent
	// input like <b><b>hi</b></b> from returning \x02\x02hi\x02\x02,
	// which wouldn't be bolded at all.
	rawName, hasAttrs := c.TagName()
	name := string(rawName)

	var bold, block bool
	var link *string

	for hasAttrs {
		var rawKey, rawValue []byte
		rawKey, rawValue, hasAttrs = c.TagAttr()
		key := string(rawKey)
		value := string(rawValue)

		switch key {
		case "value":
			switch name {
			case "button":
				link = &value
			case "input":
				c.WriteString(value)
			}

		case "href":
			if name == "a" {
				link = &value
			}

		case "alt":
			if name == "img" {
				c.WriteString(value)
			}

		// I need a better CSS parser, but that will do for now
		case "style":
			for _, rawPair := range strings.Split(value, ";") {
				pair := strings.SplitN(rawPair, ":", 2)
				if len(pair) != 2 {
					continue
				}
				property := strings.ToLower(strings.TrimSpace(pair[0]))
				argument := strings.ToLower(strings.TrimSpace(pair[1]))

				switch property {
				case "display":
					switch argument {
					case "hidden":
						c.hidden = true
					case "block", "inline-block":
						block = true
					}
				}
			}
		}
	}

	switch name {
	case "b", "strong":
		if !c.bold {
			bold = true
			c.bold = true
		}
	case "p", "td", "center", "div", "ol":
		block = true
	case "h1", "h2", "h3":
		if !c.bold {
			bold = true
			c.bold = true
		}
		block = true
	case "br", "hr":
		c.printNewline()
		return
	case "li":
		c.printNewline()
		c.WriteString("• ")
	case "img", "input":
		return
	}

	if bold {
		c.WriteByte(boldCharacter[0])
		defer c.WriteByte(boldCharacter[0])
	}
	if block {
		c.printNewline()
		defer c.printNewline()
	}
	if link != nil {
		c.WriteByte('[')
		*c.preventNewLines = true
		defer c.writeLink(*link)
	}

	c.parseToken()
}

func isNewLineCharacter(b byte) bool {
	return b == '\n' || b == boldCharacter[0]
}

func isEmpty(line string) bool {
	for _, char := range []byte(line) {
		if !isNewLineCharacter(char) {
			return false
		}
	}
	return true
}
