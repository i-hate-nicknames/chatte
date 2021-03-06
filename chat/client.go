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

// ClientHandler represents a connected remote client
// All interaction happens through the server and is written in corresponding in and out
// channels.
// Server goroutine posts commands to this client handler's command bus when it wants
// it to perform actions
// When remote client sends a message it gets stored in the in channel, that will
// eventually be read by the server
type ClientHandler struct {
	// incoming messages from the remote client, read by server goroutine
	in chan<- *protocol.Message
	// commands to this client instance to perform actions to remote client
	commandBus chan command
	ctx        context.Context
	cancel     context.CancelFunc
	Username   string
	conn       *websocket.Conn
	lastActive time.Time
}

// MakeClient with given server channel, username and connection
func MakeClient(in chan<- *protocol.Message, username string, conn *websocket.Conn) *ClientHandler {
	// todo: maybe increase buffering in command bus to avoid possible blocking in the server
	bus := make(chan command, 0)
	now := time.Now()
	return &ClientHandler{in: in, commandBus: bus, Username: username, conn: conn, lastActive: now}
}

// Start a client session. Will start reading messages from the client as well as sending
// messages back
func (c *ClientHandler) Start(ctx context.Context) {
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

// Err sends an error notification to the client
func (c *ClientHandler) Err(text string) {
	command := command{discriminator: commandError}
	command.err = &errorCommand{Err: text}
	c.sendCommand(command)
}

// Public sends a text message marked as public to the client
func (c *ClientHandler) Public(sender, text string, timestamp time.Time) {
	command := command{discriminator: commandBroadcast}
	command.broadcast = &broadcastCommand{sender: sender, text: text, timestamp: timestamp}
	c.sendCommand(command)
}

// Private sends a text message marked as private to the client
func (c *ClientHandler) Private(sender, text string, timestamp time.Time) {
	command := command{discriminator: commandPrivate}
	command.private = &privateCommand{sender: sender, text: text, timestamp: timestamp}
	c.sendCommand(command)
}

// Stop issues a stop command to this client handler.
// This will eventually result in the closing client connection
// and stopping the handler
func (c *ClientHandler) Stop() {
	command := command{discriminator: commandStop}
	c.sendCommand(command)
}

// IsRunning returns true when this handler is running
// A newly created handler starts running, and can be stopped
// by a stop command
func (c *ClientHandler) IsRunning() bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
		return true
	}
}

// send a command to the client goroutine
func (c *ClientHandler) sendCommand(cmd command) {
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
// stop client when the connection is broken
func (c *ClientHandler) readMessages() {
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

// read next message from the client connection
func (c *ClientHandler) nextMessage() (*protocol.Message, error) {
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

// handle commands posted to this client by server goroutine,
// stop client when receiving stop command
// this method also closes client connection when for any reason
// this client is stopped
func (c *ClientHandler) handleCommands() {
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

func (c *ClientHandler) handleCommand(cmd command) {
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
func (c *ClientHandler) writeRaw(data string) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(data))
	if err != nil {
		log.Println(err)
		c.cancel()
	}
}
