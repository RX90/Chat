package auth

import (
	"github.com/RX90/Chat/middleware"
	"github.com/gin-gonic/gin"
)

func (a *Auth) InitRoutes() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", a.signUp)
			auth.POST("/sign-in", a.signIn)
			auth.POST("/refresh", middleware.SoftUserIdentity, a.refreshTokens)
			auth.POST("/logout", middleware.SoftUserIdentity, a.logout)
		}
	}

	return router
}
