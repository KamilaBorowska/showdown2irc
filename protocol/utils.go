// showdown2irc - use Showdown chat with an IRC client
// Copyright (C) 2016 Konrad Borowski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package protocol

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// UserID represents a user ID. It's distinct from standard strings to
// prevent accidentally mistaking usernames with IDs
type UserID string

// RoomID represents a room ID.
type RoomID string

// SplitUser given a string with a rank and username provided User object
func SplitUser(name string) User {
	auth, size := utf8.DecodeRuneInString(name)
	return User{auth, name[size:]}
}

// ToID converts a username to its ID.
func ToID(name string) UserID {
	var buffer bytes.Buffer
	for _, character := range name {
		character = unicode.ToLower(character)
		if 'a' <= character && character <= 'z' || '0' <= character && character <= '9' {
			buffer.WriteRune(character)
		}
	}
	return UserID(buffer.String())
}

type serverDoesNotExistError struct{}

func (serverDoesNotExistError) Error() string {
	return "Server does not exist"
}

var configurationRegex = regexp.MustCompile(`(?m)^var config = (.*);$`)

type configuration struct {
	Host string
	Port uint16
}

func findConfiguration(name string) (*configuration, error) {
	if !strings.Contains(name, ".") {
		name += ".psim.us"
	}
	escapedName := url.QueryEscape(name)
	res, err := http.Get("https://play.pokemonshowdown.com/crossdomain.php?host=" + escapedName)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	matches := configurationRegex.FindSubmatch(contents)
	if matches == nil {
		return nil, new(serverDoesNotExistError)
	}
	jsonJsonData := matches[1]
	var jsonData string
	err = json.Unmarshal(jsonJsonData, &jsonData)
	if err != nil {
		return nil, err
	}
	var serverConfiguration configuration
	err = json.Unmarshal([]byte(jsonData), &serverConfiguration)
	if err != nil {
		return nil, err
	}
	if serverConfiguration.Host == "showdown" {
		// This is a default server configuration (TODO: read it from js/storage.js) instead
		serverConfiguration.Host = "sim.psim.us"
		serverConfiguration.Port = 443
	}
	return &serverConfiguration, nil
}
