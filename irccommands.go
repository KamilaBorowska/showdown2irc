package main

import (
	"fmt"
	"strings"

	"github.com/xfix/showdown2irc/protocol"
)

var ircCommands = map[string]func(*connection, []string){
	"CAP": func(c *connection, command []string) {
		// Not implemented, does nothing
	},
	"PASS": func(c *connection, command []string) {
		c.loginData.Password = command[0]
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
			c.sendNumeric(ErrNoOperHost, "No O-lines for your host")
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
			room := c.showdown.Room(protocol.RoomID(command[0][1:]))
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
		room := c.showdown.Room(protocol.RoomID(command[0][1:]))
		room.SendCommand("part", "")
	},
	"MODE": func(c *connection, command []string) {
		if len(command) == 1 {
			c.sendNumeric(RplWhoReply, command[0], "+ntc")
		}
	},
	"QUIT": func(c *connection, command []string) {
		c.close()
	},
}
