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
	"fmt"
	"strings"

	"github.com/xfix/showdown2irc/showdown"
)

var ircCommands = map[string]func(*connection, []string){
	"CAP": func(c *connection, command []string) {
		// Not implemented, does nothing
	},
	"PASS": func(c *connection, command []string) {
		if len(command) < 1 {
			c.needMoreParams("PASS")
		} else if c.userObtained || c.nickObtained {
			c.sendNumeric(ErrAlreadyRegistered)
		} else {
			c.loginData.Password = command[0]
		}
	},
	"NICK": func(c *connection, command []string) {
		if c.userObtained && !c.nickObtained {
			c.continueConnection()
		}
		c.nickObtained = true
	},
	"USER": func(c *connection, command []string) {
		c.loginData.Nickname = command[3]
		c.nickname = escapeUser(command[3])
		if !c.userObtained && c.nickObtained {
			c.continueConnection()
		}
		c.userObtained = true
	},
	"OPER": func(c *connection, command []string) {
		// The server doesn't support OPER command, so claim that the current
		// user host doesn't have O-lines, even if that's not a real issue.
		if len(command) < 2 {
			c.needMoreParams("OPER")
		} else {
			c.sendNumeric(ErrNoOperHost)
		}
	},
	"USERHOST": func(c *connection, command []string) {
		for _, arg := range command {
			c.sendNumeric(RplUserhost, escapeUserWithHost(arg))
		}
	},
	"PING": func(c *connection, command []string) {
		args := make([]string, len(command)+2)
		args[0] = "PONG"
		args[1] = "showdown"
		copy(args[2:], command)
		c.sendGlobal(args...)
	},
	"PRIVMSG": func(c *connection, command []string) {
		if command[0][0] == '#' {
			room := c.showdown.Room(showdown.RoomID(command[0][1:]))
			room.Reply(unescapeUser(command[1]))
		} else if command[1] != "NickServ" {
			c.showdown.SendGlobalCommand("pm", fmt.Sprintf("%s,%s", command[0], command[1]))
		}
	},
	"JOIN": func(c *connection, command []string) {
		for _, room := range strings.Split(command[0], ",") {
			c.showdown.SendGlobalCommand("join", room)
		}
	},
	"PART": func(c *connection, command []string) {
		room := c.showdown.Room(showdown.RoomID(command[0][1:]))
		room.SendCommand("part", "")
	},
	"MODE": func(c *connection, command []string) {
		if len(command) == 1 {
			c.sendNumeric(RplChannelModeIs, command[0], "+ntc", "")
		}
	},
	"QUIT": func(c *connection, command []string) {
		c.close()
	},
}
