package chat

import (
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
	out chan string
	// this channel is closed when remote client is disconnected
	done     chan struct{}
	username string
}

func MakeClient(in chan<- string, username string) *Client {
	out := make(chan string, 0)
	done := make(chan struct{}, 0)
	return &Client{in: in, out: out, done: done, username: username}
}

// Handle given connection, starting a client session. Will
// start reading messages from the client as well as sending
// messages back
func (c *Client) handle(conn *websocket.Conn) {
	go c.readMessages(conn)
	go c.writeMessages(conn)
}

// Read messages from the remote client, putting them on the
// in channel
func (c *Client) readMessages(conn *websocket.Conn) {
	for {
		// todo handle different message types
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		msg, err := protocol.Unmarshal(msgData)
		if err != nil {
			log.Println("Malformed message:", msg)
			continue
		}

		log.Println("Received", msg)
		c.in <- string(msg.Type)
	}
}

// Send messages, posted to this client instance to the remote client
func (c *Client) writeMessages(conn *websocket.Conn) {
	for {
		msg := <-c.out
		log.Println("Sending back to client", msg)
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			close(c.done)
			log.Println(err)
			return
		}
	}
}
