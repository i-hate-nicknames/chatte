package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	in       chan<- string
	out      chan string
	done     chan struct{}
	isActive bool
}

func MakeClient(in chan<- string) *Client {
	out := make(chan string, 0)
	done := make(chan struct{}, 0)
	return &Client{in: in, out: out, done: done}
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
			close(c.done)
			log.Println(err)
			return
		}
	}
}
