package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/xfix/showdown2irc/protocol"
)

var tokenRegexp = regexp.MustCompile(`:[^\r\n]*|[^\s:]+`)

type connection struct {
	tcp          net.Conn
	nickname     string
	showdown     *protocol.BotConnection
	loginData    protocol.LoginData
	nickObtained bool
	userObtained bool
	closing      bool
}

func (c *connection) parseIRCLine(tokens []string) {
	commandName := strings.ToUpper(tokens[0])
	if command, ok := ircCommands[commandName]; ok {
		command(c, tokens[1:])
	} else if len(tokens) >= 2 && len(tokens[1]) > 0 && tokens[1][0] == '#' {
		room := c.showdown.GetRoom(protocol.RoomID(tokens[1][1:]))
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

func (c *connection) continueConnection() {
	showdownCommands := getShowdownCommands(c)
	showdownConnection, connectionSuccess, err := protocol.ConnectToServer(c.loginData, "showdown", showdownCommands)
	if err != nil {
		c.sendGlobal("NOTICE", "#", err.Error())
		c.close()
		return
	}
	c.showdown = showdownConnection
	select {
	case <-connectionSuccess:
		c.sendGlobal("NICK", c.nickname)
		c.sendGlobal("001", c.nickname, "Welcome to Showdown proxy!")
		c.sendGlobal("005", c.nickname, "PREFIX=(qraohv)~#&@%+")
		c.sendGlobal("422", c.nickname, "No MoTD here. Go on.")
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
	"USERHOST": func(c *connection, command []string) {
		for _, arg := range command {
			c.sendGlobal("302", c.nickname, escapeUserWithHost(arg))
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
			room := c.showdown.GetRoom(protocol.RoomID(command[0][1:]))
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
		room := c.showdown.GetRoom(protocol.RoomID(command[0][1:]))
		room.SendCommand("part", "")
	},
	"MODE": func(c *connection, command []string) {
		if len(command) == 1 {
			c.sendGlobal("352", c.nickname, command[0], "+ntc")
		}
	},
	"QUIT": func(c *connection, command []string) {
		c.close()
	},
}

var rankMap = map[rune]byte{'~': 'g', '#': 'r', '&': 'a', '@': 'o', '%': 'h', '+': 'v'}

func escapeUser(name string) string {
	return strings.Replace(name, " ", "\u00A0", -1)
}

func unescapeUser(name string) string {
	return strings.Replace(name, "\u00A0", " ", -1)
}

// Some IRC clients expect host for an user during room joining operations. This generates a fake one for their purpose
func escapeUserWithHost(name string) string {
	return fmt.Sprintf("%s!%s@showdown", escapeUser(name), protocol.ToID(name))
}

func escapeRoom(room protocol.RoomID) string {
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

var whoisRegexp = regexp.MustCompile(
	`<div class="infobox"><strong class="username">` +
		`<small style="display:none">(.)</small>([^<]+)</strong> ` +
		`<br />.*Rooms: (.*)</div>`)

var roomRegexp = regexp.MustCompile(`([^\w\s]?)<a href="/([^"]+)">`)

func getShowdownCommands(c *connection) map[string]func(string, protocol.Room) {
	return map[string]func(string, protocol.Room){
		"": func(rawMessage string, room protocol.Room) {
			c.sendGlobal("NOTICE", escapeRoom(room.ID), rawMessage)
		},
		"init": func(rawMessage string, room protocol.Room) {
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
		"c:": func(rawMessage string, room protocol.Room) {
			parts := strings.SplitN(rawMessage, "|", 3)
			_, author := protocol.SplitUser(parts[1])
			escapedAuthor := escapeUser(author)
			if escapedAuthor != c.nickname {
				contents := parts[2]
				c.send(escapedAuthor, "PRIVMSG", escapeRoom(room.ID), contents)
			}
		},
		"L": func(rawMessage string, room protocol.Room) {
			_, name := protocol.SplitUser(rawMessage)
			c.send(escapeUserWithHost(name), "PART", escapeRoom(room.ID), "")
		},
		"J": func(rawMessage string, room protocol.Room) {
			rank, name := protocol.SplitUser(rawMessage)
			c.send(escapeUserWithHost(name), "JOIN", escapeRoom(room.ID))
			if ircRank, ok := rankMap[rank]; ok {
				c.sendGlobal("MODE", escapeRoom(room.ID), fmt.Sprintf("+%c", ircRank), escapeUser(name))
			}
		},
		"pm": func(rawMessage string, room protocol.Room) {
			parts := strings.SplitN(rawMessage, "|", 3)
			_, author := protocol.SplitUser(parts[0])
			contents := parts[2]
			escapedAuthor := escapeUser(author)
			if escapedAuthor != c.nickname {
				c.send(escapedAuthor, "PRIVMSG", escapedAuthor, contents)
			}
		},
		"raw": func(rawMessage string, room protocol.Room) {
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
