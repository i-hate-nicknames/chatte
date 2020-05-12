package main

import (
	"github.com/i-hate-nicknames/chatte/chat"
)

func main() {
	server := chat.MakeServer()
	go server.Run()
	chat.StartApp(server)
}
