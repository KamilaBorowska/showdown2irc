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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/xfix/showdown2irc/showdown"
)

var tokenRegexp = regexp.MustCompile(`:[^\r\n]*|[^\s:]+`)

type connection struct {
	tcp          net.Conn
	nickname     string
	showdown     *showdown.BotConnection
	loginData    showdown.LoginData
	nickObtained bool
	userObtained bool
	closing      bool
}

func (c *connection) parseIRCLine(tokens []string) {
	commandName := strings.ToUpper(tokens[0])
	if command, ok := ircCommands[commandName]; ok {
		command(c, tokens[1:])
	} else if len(tokens) >= 2 && len(tokens[1]) > 0 && tokens[1][0] == '#' {
		room := c.showdown.Room(showdown.RoomID(tokens[1][1:]))
		room.SendCommand(commandName, strings.Join(tokens[2:], ""))
	} else {
		c.showdown.SendGlobalCommand(commandName, strings.Join(tokens[1:], " "))
	}
}

func (c *connection) send(parts ...string) {
	result := toIRC(parts)
	log.Print(result)
	c.tcp.Write([]byte(result))
}

func (c *connection) sendGlobal(parts ...string) {
	newParts := make([]string, len(parts)+1)
	newParts[0] = "showdown"
	copy(newParts[1:], parts)
	c.send(newParts...)
}

func (c *connection) sendNumeric(numeric int, parts ...string) {
	newParts := make([]string, len(parts)+3)
	newParts[0] = "showdown"
	newParts[1] = fmt.Sprintf("%03d", numeric)
	newParts[2] = c.nickname
	copy(newParts[3:], parts)
	c.send(newParts...)
}

func (c *connection) needMoreParams(command string) {
	c.sendNumeric(ErrNeedMoreParams, command, "Not enough parameters")
}

func (c *connection) runShowdownCommand(command, argument string, room *showdown.Room) {
	if callback, ok := showdownCommands[command]; ok {
		callback(c, argument, room)
	}
}

func (c *connection) continueConnection() {
	showdownConnection, connectionSuccess, err := showdown.ConnectToServer(c.loginData, "showdown", c.runShowdownCommand)
	if err != nil {
		c.sendGlobal("NOTICE", "#", err.Error())
		c.close()
		return
	}
	c.showdown = showdownConnection
	select {
	case <-connectionSuccess:
		c.sendGlobal("NICK", c.nickname)
		c.sendNumeric(RplWelcome, "Welcome to Showdown proxy!")
		c.sendNumeric(RplBounce, "PREFIX=(qraohv)~#&@%+")
		c.sendNumeric(RplMOTDStart, "- showdown Message of the day - ")
		c.sendNumeric(RplMOTD, "- This server is a proxy server for Pokémon Showdown.")
		c.sendNumeric(RplMOTD, "- For source code, see https://github.com/xfix/showdown2irc")
		c.sendNumeric(RplEndOfMOTD, "End of MOTD command")
	case <-time.After(10 * time.Second):
		c.sendGlobal("NOTICE", "#", "Authentication did not succeed in 10 seconds")
		c.close()
		return
	}
}

func (c *connection) close() {
	c.sendGlobal("QUIT", c.nickname)
	c.closing = true
	if c.showdown != nil {
		c.showdown.Close()
	}
}

func escapeUser(name string) string {
	return strings.Replace(name, " ", "\u00A0", -1)
}

func unescapeUser(name string) string {
	return strings.Replace(name, "\u00A0", " ", -1)
}

// Some IRC clients expect host for an user during room joining operations. This generates a fake one for their purpose
func escapeUserWithHost(name string) string {
	return fmt.Sprintf("%s!%s@showdown", escapeUser(name), showdown.ToID(name))
}

func escapeRoom(room showdown.RoomID) string {
	return "#" + string(room)
}

func fromIRC(line string) []string {
	lines := tokenRegexp.FindAllString(line, -1)
	for i, line := range lines {
		if line[0] == ':' {
			lines[i] = line[1:]
		}
	}
	return lines
}

func toIRC(tokens []string) string {
	spaceTokenFound := false
	var result bytes.Buffer
	for i, token := range tokens {
		if spaceTokenFound {
			panic(errors.New("Tokens found after a token containing a space"))
		}
		if i == 0 {
			result.WriteByte(':')
		} else {
			result.WriteByte(' ')
			if token == "" || strings.Contains(token, " ") {
				result.WriteByte(':')
				spaceTokenFound = true
			}
		}
		result.WriteString(token)
	}
	result.WriteString("\r\n")
	return result.String()
}

func handleIRC(rawConnection net.Conn) {
	defer rawConnection.Close()
	lines := bufio.NewReader(rawConnection)
	var c connection

	c = connection{tcp: rawConnection}
	for !c.closing {
		line, err := lines.ReadString('\n')
		if err != nil {
			log.Print(err)
			return
		}
		tokens := fromIRC(line)
		log.Print(tokens)
		c.parseIRCLine(tokens)
	}
}

func listen() {
	socket, err := net.Listen("tcp", "localhost:6667")
	if err != nil {
		log.Fatal(err)
	}
	for {
		connection, err := socket.Accept()
		if err != nil {
			log.Print(err)
		}
		go handleIRC(connection)
	}
}
