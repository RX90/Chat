package router

import (
	"github.com/RX90/Chat/internal/handler"
	"github.com/RX90/Chat/web"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.Handler) (*gin.Engine, error) {
	router := gin.Default()

	tmpl, err := web.ParseTemplates()
	if err != nil {
		return nil, err
	}
	router.SetHTMLTemplate(tmpl)

	fs, err := web.StaticFiles()
	if err != nil {
		return nil, err
	}
	router.StaticFS("/static", fs)

	router.GET("/", h.HandleChatPage)
	router.GET("/ws", h.ServeWs)

	return router, nil
}
