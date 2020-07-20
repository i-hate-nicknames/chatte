package chat

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/i-hate-nicknames/chatte/protocol"
)

const inactiveTimeout = 3 * time.Second
const cleanupTime = 1 * time.Second

// Server represents a chat server that holds clients and handles
// all client interaction
type Server struct {
	// mapping of username to client handlers
	clients map[string]*ClientHandler
	mux     sync.Mutex
	// incoming messages from all clients
	in         chan *protocol.Message
	nextUserID int
}

// MakeServer creates a new instance of chat server
func MakeServer() *Server {
	in := make(chan *protocol.Message, 10)
	clients := make(map[string]*ClientHandler, 0)
	return &Server{in: in, clients: clients}
}

// Run server, listening to active websocket connections. Each new connection
// will be added,
func (s *Server) Run() {
	ticker := time.NewTicker(cleanupTime)
	for {
		select {
		case message := <-s.in:
			s.handleMessage(message)
		case <-ticker.C:
			s.mux.Lock()
			for _, client := range s.clients {
				// delete all clients that are not running
				if !client.IsRunning() {
					log.Println("Cleanup, removing", client.Username)
					delete(s.clients, client.Username)
					continue
				}
				// send stop command to goroutine with timed out client
				if time.Since(client.lastActive) > inactiveTimeout {
					log.Printf("Client %s timed out\n", client.Username)
					client.Stop()
				}
			}
			s.mux.Unlock()
		}
	}
}

// handleMessage reacts to an incoming message from a client
// It may result in sending commands to one or more client handlers
// or in some additional actions (like stopping removing client handler)
func (s *Server) handleMessage(message *protocol.Message) {
	sender, ok := s.getClient(message.Sender)
	if !ok {
		log.Printf("Message sender not found, message: %v\n", message)
		return
	}
	switch message.Discriminator {
	case protocol.TypeQuit:
		// todo: mark client for deletion
	case protocol.TypePing:
		// do nothing, ping is used just for updating client activity state
	case protocol.TypePrivate:
		recipient, ok := s.getClient(message.Private.Recipient)
		if !ok {
			msg := fmt.Sprintf("Recipient %s not found", message.Private.Recipient)
			log.Println(msg)
			sender.Err(msg)
			return
		}
		recipient.Private(sender.Username, message.Private.Text, message.Time)

	case protocol.TypePublic:
		for _, client := range s.clients {
			client.Public(sender.Username, message.Public.Text, message.Time)
		}
	}
}

func (s *Server) getClient(username string) (*ClientHandler, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	client, ok := s.clients[username]
	return client, ok
}

// handle a newly connected client: create new client handler object
// and assign it a username
func (s *Server) handleConn(ctx context.Context, conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.nextUserID++
	username := fmt.Sprintf("User%d", s.nextUserID)
	client := MakeClient(s.in, username, conn)
	client.Start(ctx)
	// todo: send hello command to the client with the username
	s.clients[username] = client
	log.Println("Connected new client:", username)
}
