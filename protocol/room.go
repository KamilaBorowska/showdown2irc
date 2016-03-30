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
	"strings"
)

// Room represents a chat room where many users can talk
type Room struct {
	Title         string
	ID            RoomID
	BotConnection *BotConnection
	UserList      map[UserID]User
}

// User represents an user name with a rank
type User struct {
	Rank rune
	Name string
}

// Reply replies to a message with a given string.
func (r *Room) Reply(message string) {
	r.BotConnection.say(message, r.ID)
}

// SendCommand sends a backslashed command to a room.
func (r *Room) SendCommand(command, value string) {
	r.BotConnection.sendCommand(command, value, r.ID)
}

func (r *Room) onUserList(userlist string) {
	r.UserList = map[UserID]User{}
	users := strings.Split(userlist, ",")
	for _, user := range users[1:] {
		r.UserList[ToID(user)] = SplitUser(user)
	}
}

func (r *Room) onJoin(username string) {
	r.UserList[ToID(username)] = SplitUser(username)
}

func (r *Room) onLeave(username string) {
	delete(r.UserList, ToID(username))
}

func (r *Room) onRename(username string, oldid UserID) {
	delete(r.UserList, oldid)
	r.onJoin(username)
}
