package chat

import (
	"context"
	"errors"
	"fmt"
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
	// outgoing messages, errors and other commands from server to this client
	commandBus chan command
	ctx        context.Context
	cancel     context.CancelFunc
	Username   string
	conn       *websocket.Conn
	lastActive time.Time
}

// MakeClient with given server channel, username and connection
func MakeClient(in chan<- *protocol.Message, username string, conn *websocket.Conn) *Client {
	bus := make(chan command, 0)
	now := time.Now()
	return &Client{in: in, commandBus: bus, Username: username, conn: conn, lastActive: now}
}

// Start a client session. Will start reading messages from the client as well as sending
// messages back
func (c *Client) Start(ctx context.Context) {
	c.ctx, c.cancel = context.WithCancel(ctx)
	go c.readMessages()
	go c.handleCommands()
}

type commandType string

const (
	commandStop      commandType = "DISCONNECT"
	commandBroadcast             = "BROADCAST"
	commandPrivate               = "PRIVATE"
	commandError                 = "ERROR"
)

type command struct {
	discriminator commandType
	broadcast     *broadcastCommand
	private       *privateCommand
	err           *errorCommand
}

type broadcastCommand struct {
	sender    string
	text      string
	timestamp time.Time
}

type privateCommand struct {
	sender    string
	text      string
	timestamp time.Time
}

type errorCommand struct {
	Err string
}

func (c *Client) Err(text string) {
	command := command{discriminator: commandError}
	command.err = &errorCommand{Err: text}
	c.sendCommand(command)
}

func (c *Client) Public(sender, text string, timestamp time.Time) {
	command := command{discriminator: commandBroadcast}
	command.broadcast = &broadcastCommand{sender: sender, text: text, timestamp: timestamp}
	c.sendCommand(command)
}

func (c *Client) Private(sender, text string, timestamp time.Time) {
	command := command{discriminator: commandPrivate}
	command.private = &privateCommand{sender: sender, text: text, timestamp: timestamp}
	c.sendCommand(command)
}

func (c *Client) Stop() {
	command := command{discriminator: commandStop}
	c.sendCommand(command)
}

func (c *Client) IsRunning() bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
		return true
	}
}

// todo: consider send command instead send message, and use send command
// for all interaction of server and client goroutines, like stopping

// send a command to the client goroutine
func (c *Client) sendCommand(cmd command) {
	// guard against cases when client is closed
	select {
	case c.commandBus <- cmd:
	case <-c.ctx.Done():
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
func (c *Client) handleCommands() {
	defer c.conn.Close()
	for {
		select {
		case cmd := <-c.commandBus:
			c.handleCommand(cmd)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) handleCommand(cmd command) {
	switch cmd.discriminator {
	case commandStop:
		log.Println("Stopping client", c.Username)
		// todo: send disconnected message to the client, with a reason
		c.cancel()
	case commandBroadcast:
		log.Println("Sending broadcast to client", cmd.broadcast.sender, cmd.broadcast.text)
		c.writeRaw(fmt.Sprintf("%s: %s", cmd.broadcast.sender, cmd.broadcast.text))
	case commandPrivate:
		log.Printf("Sending private from %s to %s, text: %s\n", cmd.private.sender, c.Username, cmd.private.text)
		c.writeRaw(fmt.Sprintf("%s: %s", cmd.private.sender, cmd.private.text))
	case commandError:
		log.Printf("Sending error %s to %s\n", cmd.err.Err, c.Username)
		c.writeRaw(fmt.Sprintf("error: %s", cmd.err.Err))
	}
}

// write raw data to the client connection, stop running this client
// in case there is an error when writing
func (c *Client) writeRaw(data string) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(data))
	if err != nil {
		log.Println(err)
		c.cancel()
	}
}
