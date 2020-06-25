package chat

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Server represents a chat server that holds clients and handles
// all client interaction
type Server struct {
	// mapping of username to client structures
	clients map[string]*Client
	mux     sync.Mutex
	// incoming messages from all clients
	in         chan *ClientMessage
	nextUserID int
}

func MakeServer() *Server {
	in := make(chan *ClientMessage, 10)
	clients := make(map[string]*Client, 0)
	return &Server{in: in, clients: clients}
}

// Run server, listening to client messages
// and broadcasting this message to all connected clients
func (s *Server) Run() {
	for {
		message := <-s.in
		for _, client := range s.clients {
			ok := client.SendMessage(string(message.Message.GetType()))
			if !ok {
				// todo: mark client for deletion
			}
		}
	}
}

// handle a newly connected client: create new client
// object and assign it a username
func (s *Server) handleConn(conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.nextUserID++
	username := fmt.Sprintf("User%d", s.nextUserID)
	client := MakeClient(s.in, username, conn)
	client.Start()
	s.clients[username] = client
}
