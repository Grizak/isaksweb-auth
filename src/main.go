package main

import (
	"log"
	"time"

	"github.com/Grizak/isaksweb-auth/src/config"
	"github.com/Grizak/isaksweb-auth/src/handlers"
	"github.com/Grizak/isaksweb-auth/src/middleware"
	"github.com/Grizak/isaksweb-auth/src/store"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	cfg := config.Load()

	if err := store.Init(cfg.DBPath); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	store.StartGC(1 * time.Hour) // Runs ever hour

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// 5 reqs/min, burst of 5 - strict for auth endpoints
	authLimiter := middleware.NewRateLimiter(rate.Every(1*time.Minute/5), 5)

	// 30 reqs/min, burst of 10 - relaxed for general endpoints
	generalLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/30), 10)

	// Public routes
	r.POST("/register", authLimiter.Middleware(), handlers.Register(cfg))
	r.POST("/oauth/token", generalLimiter.Middleware(), handlers.Token(cfg))
	r.POST("/oauth/refresh", generalLimiter.Middleware(), handlers.Refresh(cfg))
	r.POST("/oauth/revoke", authLimiter.Middleware(), handlers.Revoke(cfg))
	r.POST("/clients", authLimiter.Middleware(), handlers.RegisterClient(cfg))
	r.GET("/oauth/authorize", authLimiter.Middleware(), handlers.Authorize(cfg))
	r.POST("/oauth/authorize", generalLimiter.Middleware(), handlers.AuthorizeSubmit(cfg))

	// Protected routes
	auth := r.Group("/", middleware.RequireAuth(cfg))
	auth.GET("/me", handlers.Me)

	log.Printf("Auth server running on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
