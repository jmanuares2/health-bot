package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	dbpkg "github.com/jmanuares2/health-bot/internal/db"
	"github.com/jmanuares2/health-bot/internal/api"
)

func main() {
	ctx := context.Background()

	pool, err := dbpkg.Connect(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	queries := dbpkg.New(pool)

	// Single user — resolve by phone from env or use id=1
	userPhone := os.Getenv("USER_PHONE")
	if userPhone == "" {
		userPhone = "self"
	}
	user, err := queries.GetOrCreateUser(ctx, userPhone)
	if err != nil {
		log.Fatalf("get/create user: %v", err)
	}

	handlers := api.NewHandlers(queries, user)

	r := gin.Default()
	api.RegisterRoutes(r, handlers)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("gin run: %v", err)
	}
}
