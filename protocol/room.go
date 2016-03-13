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
		auth, name := SplitUser(user)
		r.UserList[ToID(name)] = User{auth, name}
	}
}

func (r *Room) onJoin(username string) {
	auth, name := SplitUser(username)
	r.UserList[ToID(name)] = User{auth, name}
}

func (r *Room) onLeave(username string) {
	_, name := SplitUser(username)
	delete(r.UserList, ToID(name))
}

func (r *Room) onRename(username string, oldid UserID) {
	delete(r.UserList, oldid)
	r.onJoin(username)
}
