package chat

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/i-hate-nicknames/chatte/protocol"
)

// Server represents a chat server that holds clients and handles
// all client interaction
type Server struct {
	// mapping of username to client structures
	clients map[string]*Client
	mux     sync.Mutex
	// incoming messages from all clients
	in         chan protocol.Message
	nextUserID int
}

func MakeServer() *Server {
	in := make(chan protocol.Message, 10)
	clients := make(map[string]*Client, 0)
	return &Server{in: in, clients: clients}
}

// Run server, listening to client messages
// and broadcasting this message to all connected clients
func (s *Server) Run() {
	for {
		message := <-s.in
		for _, client := range s.clients {
			s.handleMessage(client, message)
		}
	}
}

func (s *Server) handleMessage(client *Client, message protocol.Message) {
	ok := client.SendMessage(string(message.GetType()))
	if !ok {
		// todo: mark client for deletion
	}
	switch message.(type) {
	case protocol.QuitMessage:
		// todo: mark client for deletion
	case protocol.PingMessage:
		// todo: send pong
	case protocol.PrivateMessage:
		// todo: send private message to the receiver
	case protocol.PublicMessage:
		// todo: send message to every client
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
