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

package main

import (
	"bytes"
	"html"
	"regexp"
	"strings"

	"github.com/xfix/showdown2irc/html2irc"
	"github.com/xfix/showdown2irc/irc"
	"github.com/xfix/showdown2irc/showdown"
)

func parseTopic(c *connection, rawMessage string, room *showdown.Room) bool {
	const beginDescription = `<div class="infobox">The room description is: `
	const endDescription = `</div>`
	if !strings.HasPrefix(rawMessage, beginDescription) || !strings.HasSuffix(rawMessage, endDescription) {
		return false
	}

	description := rawMessage[len(beginDescription) : len(rawMessage)-len(endDescription)]
	c.sendNumeric(irc.RplTopic, escapeRoom(room.ID), html.UnescapeString(description))
	return true
}

var whoisRegexp = regexp.MustCompile(
	`<div class="infobox"><strong class="username">` +
		`<small style="display:none">(.)</small>([^<]+)</strong> ` +
		`<br />.*Rooms: (.*)</div>`)

var roomRegexp = regexp.MustCompile(`([^\w\s]?)<a href="/([^"]+)">`)

func parseWhois(c *connection, rawMessage string, room *showdown.Room) bool {
	whoisMatch := whoisRegexp.FindStringSubmatch(rawMessage)
	if whoisMatch == nil {
		return false
	}
	rank := whoisMatch[1]
	name := escapeUser(whoisMatch[2])
	rooms := whoisMatch[3]

	c.sendNumeric(irc.RplWhoisUser, name, string(showdown.ToID(name)), "showdown", "Global rank: "+rank)

	var output bytes.Buffer

	for _, room := range roomRegexp.FindAllStringSubmatch(rooms, -1) {
		rank := room[1]
		roomName := showdown.RoomID(room[2])
		output.WriteString(rank)
		output.WriteString(escapeRoom(roomName))
		// The IRC standard says that the space is after each entry, even
		// the last one. While silly, let's go with that.
		output.WriteByte(' ')
	}
	c.sendNumeric(irc.RplWhoisChannels, name, output.String())

	c.sendNumeric(irc.RplEndOfWhois, name)
	return true
}

func parseGeneric(c *connection, rawMessage string, room *showdown.Room) bool {
	// When unrecognized, use a generic parser for raw data
	for _, part := range html2irc.HTMLToIRC(rawMessage) {
		c.sendGlobal("NOTICE", escapeRoom(room.ID), part)
	}
	return true
}

var rawParsers = []func(*connection, string, *showdown.Room) bool{
	parseTopic,
	parseWhois,
	parseGeneric,
}
