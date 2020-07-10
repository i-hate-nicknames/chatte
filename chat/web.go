package chat

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StartApp(server *Server) {
	r := gin.Default()
	ctx := context.Background()
	r.LoadHTMLGlob("./dist/*")
	r.Static("/assets", "./dist")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"name": "user",
		})
	})
	// websock will upgrade client connection to websocket and
	// setup a client representation for remote client
	r.GET("/websock", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go server.handleConn(ctx, conn)
	})
	r.Run(":8080")
}
