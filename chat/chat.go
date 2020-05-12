package chat

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	in  chan<- string
	out chan string
}

type Server struct {
	clients []*Client
	mux     sync.Mutex
	in      chan string
}

func MakeClient(in chan<- string) *Client {
	out := make(chan string, 0)
	return &Client{in: in, out: out}
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
			client.out <- msg
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

func (c *Client) handle(conn *websocket.Conn) {
	go c.readMessages(conn)
	go c.writeMessages(conn)
}

func (c *Client) readMessages(conn *websocket.Conn) {
	for {
		// todo handle different message types
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Received", msg)
		c.in <- string(msg)
	}
}

func (c *Client) writeMessages(conn *websocket.Conn) {
	for {
		msg := <-c.out
		log.Println("Sending back to client", msg)
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
