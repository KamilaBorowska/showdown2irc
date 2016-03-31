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
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/xfix/showdown2irc/showdown"
)

var rankMap = map[rune]byte{'~': 'q', '#': 'r', '&': 'a', '@': 'o', '%': 'h', '+': 'v'}

var whoisRegexp = regexp.MustCompile(
	`<div class="infobox"><strong class="username">` +
		`<small style="display:none">(.)</small>([^<]+)</strong> ` +
		`<br />.*Rooms: (.*)</div>`)

var roomRegexp = regexp.MustCompile(`([^\w\s]?)<a href="/([^"]+)">`)

var showdownCommands = map[string]func(*connection, string, *showdown.Room){
	"": func(c *connection, rawMessage string, room *showdown.Room) {
		c.sendGlobal("NOTICE", escapeRoom(room.ID), rawMessage)
	},
	"init": func(c *connection, rawMessage string, room *showdown.Room) {
		room.SendCommand("roomdesc", "")
		id := escapeRoom(room.ID)
		c.send(c.nickname, "JOIN", id)
		var buffer bytes.Buffer
		for _, user := range room.UserList {
			length := buffer.Len()
			if length > 300 {
				c.sendNumeric(RplNamesReply, "@", id, buffer.String())
				buffer.Reset()
			} else if length != 0 {
				buffer.WriteByte(' ')
			}
			if user.Rank != ' ' {
				buffer.WriteRune(user.Rank)
			}
			buffer.WriteString(escapeUser(user.Name))
		}
		if buffer.Len() != 0 {
			c.sendNumeric(RplNamesReply, "@", id, buffer.String())
		}
		c.sendNumeric(RplEndOfNames, id, "End of /NAMES list.")
	},
	"c:": func(c *connection, rawMessage string, room *showdown.Room) {
		parts := strings.SplitN(rawMessage, "|", 3)
		escapedAuthor := escapeUser(showdown.SplitUser(parts[1]).Name)
		if escapedAuthor != c.nickname {
			contents := parts[2]
			c.send(escapedAuthor, "PRIVMSG", escapeRoom(room.ID), contents)
		}
	},
	"L": func(c *connection, rawMessage string, room *showdown.Room) {
		name := showdown.SplitUser(rawMessage).Name
		c.send(escapeUserWithHost(name), "PART", escapeRoom(room.ID), "")
	},
	"J": func(c *connection, rawMessage string, room *showdown.Room) {
		user := showdown.SplitUser(rawMessage)
		c.send(escapeUserWithHost(user.Name), "JOIN", escapeRoom(room.ID))
		if ircRank, ok := rankMap[user.Rank]; ok {
			c.sendGlobal("MODE", escapeRoom(room.ID), fmt.Sprintf("+%c", ircRank), escapeUser(user.Name))
		}
	},
	"pm": func(c *connection, rawMessage string, room *showdown.Room) {
		parts := strings.SplitN(rawMessage, "|", 3)
		contents := parts[2]
		escapedAuthor := escapeUser(showdown.SplitUser(parts[0]).Name)
		if escapedAuthor != c.nickname {
			c.send(escapedAuthor, "PRIVMSG", escapedAuthor, contents)
		}
	},
	"raw": func(c *connection, rawMessage string, room *showdown.Room) {
		const beginDescription = `<div class="infobox">The room description is: `
		const endDescription = `</div>`
		if strings.HasPrefix(rawMessage, beginDescription) && strings.HasSuffix(rawMessage, endDescription) {
			description := rawMessage[len(beginDescription) : len(rawMessage)-len(endDescription)]
			c.sendNumeric(RplTopic, escapeRoom(room.ID), html.UnescapeString(description))
			return
		}
		if result := whoisRegexp.FindStringSubmatch(rawMessage); result != nil {
			rank := result[1]
			name := escapeUser(result[2])
			rooms := result[3]

			c.sendNumeric(RplWhoisUser, name, string(showdown.ToID(name)), "showdown", "*", "Global rank: "+rank)

			var result bytes.Buffer

			for _, room := range roomRegexp.FindAllStringSubmatch(rooms, -1) {
				rank := room[1]
				roomName := showdown.RoomID(room[2])
				result.WriteString(rank)
				result.WriteString(escapeRoom(roomName))
				// The IRC standard says that the space is after each entry, even
				// the last one. While silly, let's go with that.
				result.WriteByte(' ')
			}
			c.sendNumeric(RplWhoisChannels, name, result.String())

			c.sendNumeric(RplEndOfWhois, name, "End of /WHOIS list")
			return
		}

		// When unrecognized, use a generic parser for raw data
		for _, part := range htmlToIRC(rawMessage) {
			c.sendGlobal("NOTICE", escapeRoom(room.ID), part)
		}
	},
}
