package chat

import (
	"context"
	"errors"
	"log"
	"time"

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
	in chan<- *protocol.Message
	// outgoing messages that should be sent to remote client
	out        chan string
	ctx        context.Context
	cancel     context.CancelFunc
	Username   string
	conn       *websocket.Conn
	lastActive time.Time
}

// MakeClient with given server channel, username and connection
func MakeClient(in chan<- *protocol.Message, username string, conn *websocket.Conn) *Client {
	out := make(chan string, 0)
	now := time.Now()
	return &Client{in: in, out: out, Username: username, conn: conn, lastActive: now}
}

// Start a client session. Will start reading messages from the client as well as sending
// messages back
func (c *Client) Start(ctx context.Context) {
	c.ctx, c.cancel = context.WithCancel(ctx)
	go c.readMessages()
	go c.writeMessages()
}

func (c *Client) Stop() {
	c.cancel()
}

// todo: consider send command instead send message, and use send command
// for all interaction of server and client goroutines, like stopping

// SendMessage sends message to the client. Return running status
// of the client at the moment of running. The return status signifies that
// the message has been scheduled, not the delivery status. Disconnected client
// will eventually start to return false upon calling this method
// This method is supposed to be called from the server goroutine
// todo: change string to some message type
func (c *Client) SendMessage(message string) bool {
	select {
	case c.out <- message:
		return true
	case <-c.ctx.Done():
		return false
	}
}

type errMalformedMessage struct {
	err error
}

func (e *errMalformedMessage) Error() string {
	return "Malformed message: " + e.err.Error()
}

func (e *errMalformedMessage) Unwrap() error {
	return e.err
}

// Read messages from the remote client and put on the server's incoming
// messages channel
func (c *Client) readMessages() {
	for {

		message, err := c.nextMessage()
		var e *errMalformedMessage
		if errors.As(err, &e) {
			log.Println("Malformed message:", message, "error:", err)
			continue
		}
		if err != nil {
			select {
			case <-c.ctx.Done():
				// reading from closed connection because client should
				// be stopped
				return
			default:
				log.Println(err)
				c.cancel()
			}
			return
		}
		message.Time = time.Now()
		message.Sender = c.Username
		log.Println("Recieved", message)
		// pass message to the server goroutine
		c.in <- message
		c.lastActive = message.Time
	}
}

func (c *Client) nextMessage() (*protocol.Message, error) {
	_, messageData, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	message, err := protocol.Unmarshal(messageData)
	if err != nil {
		return nil, &errMalformedMessage{err: err}
	}
	return message, nil
}

// Read messages posted to this client instance and send them to the remote client
func (c *Client) writeMessages() {
	defer c.conn.Close()
	for {
		select {
		case message := <-c.out:
			log.Println("Sending back to client", message)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println(err)
				c.cancel()
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}
