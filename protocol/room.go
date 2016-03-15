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
