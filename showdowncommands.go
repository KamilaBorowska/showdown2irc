package main

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/xfix/showdown2irc/protocol"
)

var whoisRegexp = regexp.MustCompile(
	`<div class="infobox"><strong class="username">` +
		`<small style="display:none">(.)</small>([^<]+)</strong> ` +
		`<br />.*Rooms: (.*)</div>`)

var roomRegexp = regexp.MustCompile(`([^\w\s]?)<a href="/([^"]+)">`)

var showdownCommands = map[string]func(*connection, string, protocol.Room){
	"": func(c *connection, rawMessage string, room protocol.Room) {
		c.sendGlobal("NOTICE", escapeRoom(room.ID), rawMessage)
	},
	"init": func(c *connection, rawMessage string, room protocol.Room) {
		room.SendCommand("roomdesc", "")
		id := escapeRoom(room.ID)
		c.send(c.nickname, "JOIN", id)
		var buffer bytes.Buffer
		for _, user := range room.UserList {
			length := buffer.Len()
			if length > 300 {
				c.sendGlobal("353", c.nickname, "@", id, buffer.String())
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
			c.sendGlobal("353", c.nickname, "@", id, buffer.String())
		}
		c.sendGlobal("366", c.nickname, id, "End of /NAMES list.")
	},
	"c:": func(c *connection, rawMessage string, room protocol.Room) {
		parts := strings.SplitN(rawMessage, "|", 3)
		_, author := protocol.SplitUser(parts[1])
		escapedAuthor := escapeUser(author)
		if escapedAuthor != c.nickname {
			contents := parts[2]
			c.send(escapedAuthor, "PRIVMSG", escapeRoom(room.ID), contents)
		}
	},
	"L": func(c *connection, rawMessage string, room protocol.Room) {
		_, name := protocol.SplitUser(rawMessage)
		c.send(escapeUserWithHost(name), "PART", escapeRoom(room.ID), "")
	},
	"J": func(c *connection, rawMessage string, room protocol.Room) {
		rank, name := protocol.SplitUser(rawMessage)
		c.send(escapeUserWithHost(name), "JOIN", escapeRoom(room.ID))
		if ircRank, ok := rankMap[rank]; ok {
			c.sendGlobal("MODE", escapeRoom(room.ID), fmt.Sprintf("+%c", ircRank), escapeUser(name))
		}
	},
	"pm": func(c *connection, rawMessage string, room protocol.Room) {
		parts := strings.SplitN(rawMessage, "|", 3)
		_, author := protocol.SplitUser(parts[0])
		contents := parts[2]
		escapedAuthor := escapeUser(author)
		if escapedAuthor != c.nickname {
			c.send(escapedAuthor, "PRIVMSG", escapedAuthor, contents)
		}
	},
	"raw": func(c *connection, rawMessage string, room protocol.Room) {
		const beginDescription = `<div class="infobox">The room description is: `
		const endDescription = `</div>`
		if strings.HasPrefix(rawMessage, beginDescription) && strings.HasSuffix(rawMessage, endDescription) {
			description := rawMessage[len(beginDescription) : len(rawMessage)-len(endDescription)]
			c.sendGlobal("332", c.nickname, escapeRoom(room.ID), html.UnescapeString(description))
			return
		}
		if result := whoisRegexp.FindStringSubmatch(rawMessage); result != nil {
			rank := result[1]
			name := escapeUser(result[2])
			rooms := result[3]

			c.sendGlobal("311", c.nickname, name, string(protocol.ToID(name)), "showdown", "*", "Global rank: "+rank)

			var result bytes.Buffer

			for _, room := range roomRegexp.FindAllStringSubmatch(rooms, -1) {
				rank := room[1]
				roomName := protocol.RoomID(room[2])
				result.WriteString(rank)
				result.WriteString(escapeRoom(roomName))
				// The IRC standard says that the space is after each entry, even
				// the last one. While silly, let's go with that.
				result.WriteByte(' ')
			}
			c.sendGlobal("319", c.nickname, name, result.String())

			c.sendGlobal("318", c.nickname, name, "End of /WHOIS list")
			return
		}

		// When unrecognized, use a generic parser for raw data
		for _, part := range htmlToIRC(rawMessage) {
			c.sendGlobal("NOTICE", escapeRoom(room.ID), part)
		}
	},
}
