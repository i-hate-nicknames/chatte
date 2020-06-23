package chat

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/i-hate-nicknames/chatte/protocol"
)

// Client represents a connected remote client
// All interaction happens through the server and is written in corresponding in and out
// channels.
// When server needs to send the remote client something it puts message on
// the client out channel
// When remote client sends a message it gets stored in the in channel, that will
// eventually be read by the server
type Client struct {
	// incoming messages from the remote client
	in chan<- string
	// outgoing messages that should be sent to remote client
	out      chan string
	done     chan struct{}
	username string
	conn     *websocket.Conn
}

// MakeClient with given server channel, username and connection
func MakeClient(in chan<- string, username string, conn *websocket.Conn) *Client {
	out := make(chan string, 0)
	done := make(chan struct{}, 0)
	return &Client{in: in, out: out, username: username, conn: conn, done: done}
}

// Start a client session. Will start reading messages from the client as well as sending
// messages back
func (c *Client) Start() {
	go c.readMessages()
	go c.writeMessages()
}

// SendMessage sends message to the client. Return running status
// of the client at the moment of running. The return status signifies that
// the message has been scheduled, not the delivery status. Disconnected client
// will eventually start to return false upon calling this method
// This method is supposed to be called from the server goroutine
func (c *Client) SendMessage(message string) bool {
	select {
	case c.out <- message:
		return true
	case <-c.done:
		return false
	}
}

var errMalformedMessage = errors.New("Malformed message")

// Read messages from the remote client and put on the server's incoming
// messages channel
func (c *Client) readMessages() {
	for {
		msg, err := c.nextMessage()
		if err != nil && errors.Is(err, errMalformedMessage) {
			log.Println("Malformed message:", msg)
			continue
		}
		if err != nil {
			log.Println(err)
			close(c.done)
			return
		}
		log.Println("Received", msg)
		c.in <- string(msg.Type)
	}
}

func (c *Client) nextMessage() (*protocol.Message, error) {
	_, msgData, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	msg, err := protocol.Unmarshal(msgData)
	if err != nil {
		log.Println("Malformed message:", msg)
		return nil, errMalformedMessage
	}
	return msg, nil
}

// Read messages posted to this client instance and send them to the remote client
func (c *Client) writeMessages() {
	for {
		select {
		case msg := <-c.out:
			log.Println("Sending back to client", msg)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println(err)
				close(c.done)
				return
			}
		case <-c.done:
			return
		}
	}
}
