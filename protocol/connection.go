package protocol

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type writerMessage struct {
	messageType int
	contents []byte
}

type connection struct {
	websocket      *websocket.Conn
	finished       chan struct{}
	writer         chan writerMessage
	messageChannel chan string
}

// Closes a connection with a web socket
func (c *connection) Close() error {
	c.writer <- writerMessage{websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")}

	select {
	case <-c.finished:
	case <-time.After(time.Second):
	}

	c.websocket.Close()

	return nil
}

func (c *connection) write(message string) {
	c.writer <- writerMessage{websocket.TextMessage, []byte(message)}
}

func (c *connection) startWriter() {
	for message := range c.writer {
		c.websocket.WriteMessage(message.messageType, message.contents)
		time.Sleep(400 * time.Millisecond)
	}
}

func (c *connection) startReader() {
	for {
		messageType, message, err := c.websocket.ReadMessage()
		if err != nil {
			log.Print(err)
			c.finished <- struct{}{}
			close(c.writer)
			return
		}

		if messageType != websocket.TextMessage {
			log.Print("Unexpected message:", message)
			c.Close()
			continue
		}

		c.messageChannel <- string(message)
	}
}

const httpsPort = 443

func webSocketConnect(config *configuration) (*connection, error) {
	scheme := "ws"
	if config.Port == httpsPort {
		scheme = "wss"
	}
	url := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", config.Host, config.Port),
		Path:   "/showdown/websocket",
	}
	websocket, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	socketConnection := connection{
		websocket:      websocket,
		writer:         make(chan writerMessage),
		messageChannel: make(chan string),
		finished:       make(chan struct{}, 1),
	}

	go socketConnection.startWriter()
	go socketConnection.startReader()

	return &socketConnection, nil
}
