package server

import (
	"net/http"

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

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})

	router.GET("/ws", h.Chat.ServeWS)

	auth := router.Group("/auth")
	{
		auth.GET("/sign-up", func(c *gin.Context) {
			c.HTML(http.StatusOK, "sign-up.html", nil)
		})
		auth.GET("/sign-in", func(c *gin.Context) {
			c.HTML(http.StatusOK, "sign-in.html", nil)
		})
	}

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", h.Auth.SignUp)
			auth.POST("/sign-in", h.Auth.SignIn)
		}
	}

	return router, nil
}