package router

import (
	"net/http"

	"github.com/RX90/Chat/internal/ws"
	"github.com/RX90/Chat/web"
	"github.com/gin-gonic/gin"
)

func NewRouter(hub *ws.Hub) *gin.Engine {
	router := gin.Default()

	tmpl := web.ParseTemplates()
	router.SetHTMLTemplate(tmpl)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/ws", func(c *gin.Context) {
		ws.ServeWs(hub, c.Writer, c.Request)
	})

	return router
}
