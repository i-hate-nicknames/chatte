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

// quick hack to allow different origins when running both parcel dev server
// and go web server
var allowedHosts = []string{"http://localhost:1234", "http://localhost:8080"}

func StartApp(server *Server) {
	r := gin.Default()
	ctx := context.Background()
	upgrader.CheckOrigin = func(req *http.Request) bool {
		reqOrigin := req.Header.Get("Origin")
		if reqOrigin == "" {
			return false
		}
		for _, host := range allowedHosts {
			if host == reqOrigin {
				return true
			}
		}
		return false
	}
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
