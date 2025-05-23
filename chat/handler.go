package chat

import (
	"github.com/RX90/Chat/middleware"
	"github.com/gin-gonic/gin"
)

func (c *Chat) InitRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/chat", c.handleChatPage)
	router.GET("/ws", middleware.StrictUserIdentity, c.handleChatWS)

	return router
}
