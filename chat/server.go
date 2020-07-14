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

const inactiveTimeout = 200 * time.Second

// Server represents a chat server that holds clients and handles
// all client interaction
type Server struct {
	// mapping of username to client structures
	clients  map[string]*Client
	toDelete []string
	mux      sync.Mutex
	// incoming messages from all clients
	in         chan *protocol.Message
	nextUserID int
}

func MakeServer() *Server {
	in := make(chan *protocol.Message, 10)
	clients := make(map[string]*Client, 0)
	toDelete := make([]string, 0)
	return &Server{in: in, clients: clients, toDelete: toDelete}
}

// Run server, listening to client messages
// and broadcasting this message to all connected clients
func (s *Server) Run() {
	ticker := time.NewTicker(inactiveTimeout)
	for {
		select {
		case message := <-s.in:
			s.handleMessage(message)
		case <-ticker.C:
			for _, client := range s.clients {
				if time.Since(client.lastActive) > inactiveTimeout {
					log.Printf("Client %s timed out\n", client.Username)
					client.Stop()
				}
			}
		}

		// todo: clients are stopped, just remove them and get rid of toDelete
		for _, name := range s.toDelete {
			s.mux.Lock()
			log.Println("Disconnecting", name)
			client := s.clients[name]
			client.Stop()
			delete(s.clients, name)
			s.mux.Unlock()
		}
		s.toDelete = s.toDelete[:0]
	}
}

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
			log.Println("Recipient not found", message.Private.Recipient)
			return
		}
		recipient.Private(sender.Username, message.Private.Text, message.Time)

	case protocol.TypePulic:
		for _, client := range s.clients {
			client.Public(sender.Username, message.Public.Text, message.Time)
		}
	}
}

func (s *Server) getClient(username string) (*Client, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	client, ok := s.clients[username]
	return client, ok
}

// handle a newly connected client: create new client
// object and assign it a username
func (s *Server) handleConn(ctx context.Context, conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.nextUserID++
	username := fmt.Sprintf("User%d", s.nextUserID)
	client := MakeClient(s.in, username, conn)
	client.Start(ctx)
	s.clients[username] = client
}
