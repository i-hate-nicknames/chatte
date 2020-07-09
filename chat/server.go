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

// Server represents a chat server that holds clients and handles
// all client interaction
type Server struct {
	// mapping of username to client structures
	clients  map[string]*Client
	toDelete []string
	mux      sync.Mutex
	// incoming messages from all clients
	in         chan protocol.Message
	nextUserID int
}

func MakeServer() *Server {
	in := make(chan protocol.Message, 10)
	clients := make(map[string]*Client, 0)
	toDelete := make([]string, 0)
	return &Server{in: in, clients: clients, toDelete: toDelete}
}

// Run server, listening to client messages
// and broadcasting this message to all connected clients
func (s *Server) Run() {
	for {
		message := <-s.in
		for _, client := range s.clients {
			s.handleMessage(client, message)
		}
		for _, toDelete := range s.toDelete {
			client := s.clients[toDelete]
			client.Stop()
			delete(s.clients, toDelete)
		}
	}
}

func (s *Server) handleMessage(client *Client, message protocol.Message) {
	if time.Since(client.lastActive) > inactiveTimeout {
		s.toDelete = append(s.toDelete, client.Username)
		log.Printf("Client %s disconnected\n", client.Username)
		return
	}
	switch message.(type) {
	case protocol.QuitMessage:
		// todo: mark client for deletion
	case protocol.PingMessage:
		// do nothing, ping is used just for updating client activity state
	case protocol.PrivateMessage:
		// todo: send private message to the receiver
	case protocol.PublicMessage:
		// todo: send message to every client
	}

	// todo: remove, sending type for testing for now
	ok := client.SendMessage(string(message.GetType()))
	if !ok {
		// todo: mark client for deletion
	}
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
