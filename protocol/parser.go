package protocol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// LoginData represents authentication information for a bot
type LoginData struct {
	Nickname string
	Password string
	Rooms    []string
}

// BotConnection represents a websocket communication with a bot
type BotConnection struct {
	loginData       LoginData
	rooms           map[RoomID]Room
	commandCallback func(command, argument string, room Room)
	onSuccess       chan<- struct{}
	*connection
}

func (bc *BotConnection) Room(id RoomID) Room {
	if roomWithUsers, ok := bc.rooms[id]; ok {
		return roomWithUsers
	} else {
		return Room{BotConnection: bc, ID: id}
	}
}

func (bc *BotConnection) handleMessage(message string) {
	log.Println(message)
	var roomID RoomID
	if message[0] == '>' {
		parts := strings.SplitN(message[1:], "\n", 2)
		roomID = RoomID(parts[0])
		message = parts[1]
	}
	if message[0] == '|' {
		parts := strings.SplitN(message[1:], "|", 2)
		command := parts[0]
		var argument string
		if len(parts) > 1 {
			argument = parts[1]
		}

		if handler, ok := serverCommandHandlers[command]; ok {
			handler(argument, bc.Room(roomID))
		}
		bc.commandCallback(command, argument, bc.Room(roomID))
	} else {
		bc.commandCallback("", message, bc.Room(roomID))
	}
}

// SendGlobalCommand sends a command does not care about a room in which
// it is used.
func (bc *BotConnection) SendGlobalCommand(command string, value string) {
	bc.sendCommand(command, value, "")
}

// SendCommand uses a command in a room.
func (bc *BotConnection) sendCommand(command string, value string, room RoomID) {
	if room == "lobby" {
		room = ""
	}
	bc.write(fmt.Sprintf("%s|/%s %s", room, command, value))
}

// Say says a text message in a specified room.
func (bc *BotConnection) say(message string, room RoomID) {
	if message == "" {
		return
	}
	if message[0] == '/' {
		message = "/" + message
	} else if message[0] == '!' || strings.HasPrefix(message, ">> ") || strings.HasPrefix(message, ">>> ") {
		message = " " + message
	}
	bc.write(fmt.Sprintf("%s|%s", room, message))
}

func handleConnection(botConnection *BotConnection) {
	for message := range botConnection.messageChannel {
		botConnection.handleMessage(message)
	}
}

// ConnectToServer connects to a Showdown server by using its
// client location or its name.
func ConnectToServer(loginData LoginData, name string, commandCallback func(command, argument string, room Room)) (*BotConnection, <-chan struct{}, error) {
	conf, err := findConfiguration(name)
	if err != nil {
		return nil, nil, err
	}
	connection, err := webSocketConnect(conf)
	if err != nil {
		return nil, nil, err
	}
	onSuccess := make(chan struct{}, 1)
	botConnection := &BotConnection{
		connection:      connection,
		loginData:       loginData,
		rooms:           make(map[RoomID]Room),
		commandCallback: commandCallback,
		onSuccess:       onSuccess,
	}
	go handleConnection(botConnection)
	return botConnection, onSuccess, nil
}

var serverCommandHandlers = map[string]func(string, Room){
	"challstr": challStr,
	"init":     initializeChatRoom,
	"j":        joinRoom,
	"J":        joinRoom,
	"l":        leaveRoom,
	"L":        leaveRoom,
	"N":        renameNick,
}

const actionURL = "https://play.pokemonshowdown.com/action.php"

func challStr(challenge string, room Room) {
	botConnection := room.BotConnection
	loginData := botConnection.loginData

	parameters := url.Values{}
	parameters.Set("act", "login")
	parameters.Set("name", loginData.Nickname)
	parameters.Set("pass", loginData.Password)
	parameters.Set("challstr", challenge)

	res, err := http.PostForm(actionURL, parameters)
	if err != nil {
		log.Fatal(err)
	}

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()

	var assertion struct {
		Assertion string
	}
	json.Unmarshal(contents[1:], &assertion)

	value := fmt.Sprintf("%s,0,%s", loginData.Nickname, assertion.Assertion)
	botConnection.SendGlobalCommand("trn", value)

	botConnection.onSuccess <- struct{}{}

	for _, room := range loginData.Rooms {
		botConnection.SendGlobalCommand("join", room)
	}
}

func initializeChatRoom(rawMessage string, room Room) {
	logs := strings.Split(rawMessage, "\n")
	for _, message := range logs {
		titleMessage := "|title|"
		userListMessage := "|users|"
		if strings.HasPrefix(message, titleMessage) {
			room.Title = message[len(titleMessage):]
		} else if strings.HasPrefix(message, userListMessage) {
			room.onUserList(message[len(userListMessage):])
		}
	}
	room.BotConnection.rooms[room.ID] = room
}

func joinRoom(rawMessage string, room Room) {
	room.onJoin(rawMessage)
}

func leaveRoom(rawMessage string, room Room) {
	room.onLeave(rawMessage)
}

func renameNick(rawMessage string, room Room) {
	parts := strings.SplitN(rawMessage, "|", 2)
	room.onRename(parts[0], UserID(parts[1]))
}
