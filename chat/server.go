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
	ticker := time.NewTicker(inactiveTimeout)
	for {
		select {
		case message := <-s.in:
			s.handleMessage(message)
		case <-ticker.C:
			for _, client := range s.clients {
				if time.Since(client.lastActive) > inactiveTimeout {
					s.toDelete = append(s.toDelete, client.Username)
					log.Printf("Client %s timed out\n", client.Username)
				}
			}
		}
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

func (s *Server) handleMessage(messageUntyped protocol.Message) {
	switch message := messageUntyped.(type) {
	case protocol.QuitMessage:
		// todo: mark client for deletion
	case protocol.PingMessage:
		// do nothing, ping is used just for updating client activity state
	case protocol.PrivateMessage:
		s.mux.Lock()
		recipient, ok := s.clients[message.Recipient]
		s.mux.Unlock()
		if !ok {
			log.Println("Recipient not found", message.Recipient)
			return
		}
		recipient.SendMessage(fmt.Sprintf("private message: %s", message.Text))

	case protocol.PublicMessage:
		for _, client := range s.clients {
			ok := client.SendMessage(fmt.Sprintf("%s: %s", client.Username, message.Text))
			if !ok {
				s.toDelete = append(s.toDelete, client.Username)
			}
		}
	}

	// todo: remove, sending type for testing for now

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
