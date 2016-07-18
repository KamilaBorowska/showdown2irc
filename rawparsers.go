package main

import (
	"bytes"
	"html"
	"regexp"
	"strings"

	"github.com/xfix/showdown2irc/showdown"
)

func parseTopic(c *connection, rawMessage string, room *showdown.Room) bool {
	const beginDescription = `<div class="infobox">The room description is: `
	const endDescription = `</div>`
	if !strings.HasPrefix(rawMessage, beginDescription) || !strings.HasSuffix(rawMessage, endDescription) {
		return false
	}

	description := rawMessage[len(beginDescription) : len(rawMessage)-len(endDescription)]
	c.sendNumeric(RplTopic, escapeRoom(room.ID), html.UnescapeString(description))
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

	c.sendNumeric(RplWhoisUser, name, string(showdown.ToID(name)), "showdown", "Global rank: "+rank)

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
	c.sendNumeric(RplWhoisChannels, name, output.String())

	c.sendNumeric(RplEndOfWhois, name)
	return true
}

func parseGeneric(c *connection, rawMessage string, room *showdown.Room) bool {
	// When unrecognized, use a generic parser for raw data
	for _, part := range htmlToIRC(rawMessage) {
		c.sendGlobal("NOTICE", escapeRoom(room.ID), part)
	}
	return true
}

var rawParsers = []func(*connection, string, *showdown.Room) bool{
	parseTopic,
	parseWhois,
	parseGeneric,
}
