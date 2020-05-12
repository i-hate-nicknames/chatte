package chat

import (
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
	r.LoadHTMLGlob("static/templates/*")
	r.Static("/assets", "./static/assets")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"name": "user",
		})
	})
	r.GET("/websock", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go server.handleConn(conn)
	})
	r.Run(":8080")
}
