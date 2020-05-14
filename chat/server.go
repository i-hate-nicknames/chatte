package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	clients []*Client
	mux     sync.Mutex
	in      chan string
}

func MakeServer() *Server {
	in := make(chan string, 10)
	clients := make([]*Client, 0)
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
	client := MakeClient(s.in)
	client.handle(conn)
	s.clients = append(s.clients, client)
}
