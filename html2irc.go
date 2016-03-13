package main

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type htmlConverter struct {
	*bytes.Buffer
	*html.Tokenizer
	block        *bool
	bold, hidden bool
}

func htmlToIRC(code string) []string {
	converter := htmlConverter{
		Buffer:    new(bytes.Buffer),
		Tokenizer: html.NewTokenizer(strings.NewReader(code)),
		block:     new(bool),
	}
	converter.parseToken()
	var result []string
	for _, line := range strings.Split(converter.String(), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
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
				*c.block = false
				text := c.Text()
				c.Write(text)
			}

		case html.EndTagToken, html.ErrorToken:
			return
		}
	}
}

func (c htmlConverter) parseStartToken() {
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
						if !*c.block {
							block = true
							*c.block = true
						}
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
		if !*c.block {
			block = true
			*c.block = true
		}
	case "h1", "h2", "h3":
		if !c.bold {
			bold = true
			c.bold = true
		}
		if !*c.block {
			block = true
			*c.block = true
		}
	case "br", "hr":
		c.WriteByte('\n')
		if c.bold {
			c.WriteByte('\x02')
		}
		return
	case "li":
		c.WriteString("\n• ")
	case "img", "input":
		return
	}

	if block {
		c.WriteByte('\n')
		defer c.WriteByte('\n')
	}
	if bold {
		c.WriteByte('\x02')
		defer c.WriteByte('\x02')
	}
	if link != nil {
		c.WriteByte('[')
		defer c.WriteString(fmt.Sprintf("](%s)", *link))
	}

	if name != "img" {
		c.parseToken()
	}
}
