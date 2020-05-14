package chat

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	clients    map[string]*Client
	mux        sync.Mutex
	in         chan string
	nextUserID int
}

func MakeServer() *Server {
	in := make(chan string, 10)
	clients := make(map[string]*Client, 0)
	return &Server{in: in, clients: clients}
}

func (s *Server) Run() {
	for {
		msg := <-s.in
		for _, client := range s.clients {
			select {
			case client.out <- msg:
				continue
			case <-client.done:
				// todo: mark client for deletion
				continue
			}
		}
	}
}

func (s *Server) handleConn(conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.nextUserID++
	username := fmt.Sprintf("User%d", s.nextUserID)
	client := MakeClient(s.in, username)
	client.handle(conn)
	s.clients[username] = client
}
