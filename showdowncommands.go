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
	"strings"

	"github.com/xfix/showdown2irc/irc"
	"github.com/xfix/showdown2irc/showdown"
)

var rankMap = map[rune]byte{'~': 'q', '#': 'r', '&': 'a', '@': 'o', '%': 'h', '*': 'B', '+': 'v'}

func meCallback(c *connection, argument string, author string, room *showdown.Room) {
	c.send(author, "PRIVMSG", escapeRoom(room.ID), fmt.Sprintf("\x01ACTION %s\x01", argument))
}

var chatMessageCallbacks = map[string]func(*connection, string, string, *showdown.Room){
	"me":  meCallback,
	"mee": meCallback,
}

var showdownCommands = map[string]func(*connection, string, *showdown.Room){
	"": func(c *connection, rawMessage string, room *showdown.Room) {
		c.sendGlobal("NOTICE", escapeRoom(room.ID), rawMessage)
	},
	"users": func(c *connection, rawMessage string, room *showdown.Room) {
		room.SendCommand("roomdesc", "")
		id := escapeRoom(room.ID)
		c.send(c.nickname, "JOIN", id)
		var buffer bytes.Buffer
		for _, user := range room.UserList {
			length := buffer.Len()
			if length > 300 {
				c.sendNumeric(irc.RplNamesReply, '=', id, buffer.String())
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
			c.sendNumeric(irc.RplNamesReply, '=', id, buffer.String())
		}
		c.sendNumeric(irc.RplEndOfNames, id)
	},
	"c:": func(c *connection, rawMessage string, room *showdown.Room) {
		parts := strings.SplitN(rawMessage, "|", 3)
		escapedAuthor := escapeUserWithHost(showdown.SplitUser(parts[1]).Name)
		contents := parts[2]
		if strings.HasPrefix(contents, "//") {
			// Get rid of one /
			contents = contents[1:]
		} else if strings.HasPrefix(contents, "/") {
			parts := strings.SplitN(contents[1:], " ", 2)
			command := parts[0]
			argument := ""
			if len(parts) == 2 {
				argument = parts[1]
			}
			if callback, ok := chatMessageCallbacks[command]; ok {
				callback(c, argument, escapedAuthor, room)
				return
			}
		}
		if escapedAuthor != c.nickname {
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
		escapedAuthor := escapeUserWithHost(showdown.SplitUser(parts[0]).Name)
		if escapedAuthor != c.nickname {
			c.send(escapedAuthor, "PRIVMSG", escapedAuthor, contents)
		}
	},
	"raw":  htmlCommand,
	"html": htmlCommand,
}

func htmlCommand(c *connection, rawMessage string, room *showdown.Room) {
	// This works by trying to use each parser on a raw result, hoping
	// that one will match a pattern. This is done, because some raw
	// outputs have to be parsed it specific way in order to better
	// match IRC commands.
	for _, parser := range rawParsers {
		if parser(c, rawMessage, room) {
			return
		}
	}
}
