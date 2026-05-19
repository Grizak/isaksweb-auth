package main

import (
	"log"

	"github.com/Grizak/isaksweb-auth/src/config"
	"github.com/Grizak/isaksweb-auth/src/handlers"
	"github.com/Grizak/isaksweb-auth/src/middleware"
	"github.com/Grizak/isaksweb-auth/src/store"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := store.Init(cfg.DBPath); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Public routes
	r.POST("/register", handlers.Register(cfg))
	r.POST("/oauth/token", handlers.Token(cfg))

	// Protected routes
	auth := r.Group("/", middleware.RequireAuth(cfg))
	auth.GET("/me", handlers.Me)

	log.Printf("Auth server running on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
