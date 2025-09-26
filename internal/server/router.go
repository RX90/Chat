package server

import (
	"time"

	"github.com/RX90/Chat/config"
	"github.com/RX90/Chat/internal/handler"
	"github.com/RX90/Chat/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.Handler, cfg *config.CORSConfig) (*gin.Engine, error) {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge),
	}))

	router.Use(func(c *gin.Context) {
		c.Header("Content-Security-Policy",
				 "default-src 'self'; " +
				 "script-src 'self'; " +
				 "style-src 'self'; " +
				 "img-src 'self'; " +
				 "connect-src 'self'; " +
				 "frame-ancestors 'none';",
		)
		c.Next()
	})

	router.GET("/ws", middleware.StrictUserIdentity, h.Chat.ServeWS)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", h.Auth.SignUp)
			auth.POST("/sign-in", h.Auth.SignIn)
			auth.POST("/sign-out", middleware.SoftUserIdentity, h.Auth.SignOut)
			auth.POST("/refresh", middleware.SoftUserIdentity, h.Auth.Refresh)
			auth.POST("/verify", middleware.StrictUserIdentity, h.Auth.Verify)
		}
	}

	return router, nil
}
