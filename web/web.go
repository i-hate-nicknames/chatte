package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartApp() {
	r := gin.Default()
	r.LoadHTMLGlob("static/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"name": "user",
		})
	})
	r.Run(":8080")
}
