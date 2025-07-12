package router

import (
	"github.com/RX90/Chat/internal/handler"
	"github.com/RX90/Chat/web"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.Handler) *gin.Engine {
	router := gin.Default()

	tmpl := web.ParseTemplates()
	router.SetHTMLTemplate(tmpl)
	router.StaticFS("/static", web.StaticFiles())

	router.GET("/", h.HandleChatPage)
	router.GET("/ws", h.ServeWs)

	return router
}
